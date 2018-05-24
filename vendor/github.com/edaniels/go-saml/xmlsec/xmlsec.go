// Dervied from:
//		https://github.com/crewjam/go-xmlsec
// 		https://github.com/RobotsAndPencils/go-saml
package xmlsec

import (
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	xmlAssertionID = "urn:oasis:names:tc:SAML:2.0:assertion:Assertion"
	xmlResponseID  = "urn:oasis:names:tc:SAML:2.0:protocol:Response"
	xmlRequestID   = "urn:oasis:names:tc:SAML:2.0:protocol:AuthnRequest"
)

// SignRequest sign a SAML 2.0 AuthnRequest
func SignRequest(xml string, privateKey string) (string, error) {
	return sign(xml, privateKey, xmlRequestID)
}

// SignResponse sign a SAML 2.0 Response
func SignResponse(xml string, privateKey string) (string, error) {
	return sign(xml, privateKey, xmlResponseID)
}

// SignRaw sign plain xml
func SignRaw(xml string, privateKey string) (string, error) {
	return sign(xml, privateKey, "")
}

func sign(xml string, privateKey string, id string) (string, error) {

	privateKeyFile, err := writeToTemp(privateKey)
	if err != nil {
		return "", err
	}
	defer deleteTempFile(privateKeyFile.Name())

	samlXmlsecInput, err := writeToTemp(xml)
	if err != nil {
		return "", err
	}
	defer deleteTempFile(samlXmlsecInput.Name())

	samlXmlsecOutput, err := ioutil.TempFile(os.TempDir(), "tmpgs")
	if err != nil {
		return "", err
	}
	defer deleteTempFile(samlXmlsecOutput.Name())
	samlXmlsecOutput.Close()

	args := []string{
		"--sign", "--privkey-pem", privateKeyFile.Name(), "--output", samlXmlsecOutput.Name(),
	}
	if len(id) != 0 {
		args = append(args, "--id-attr:ID", id)
	}
	args = append(args, samlXmlsecInput.Name())
	output, err := exec.Command("xmlsec1", args...).CombinedOutput()
	if err != nil {
		return "", errors.New(err.Error() + " : " + string(output))
	}

	samlSignedRequest, err := ioutil.ReadFile(samlXmlsecOutput.Name())
	if err != nil {
		return "", err
	}
	samlSignedRequestXML := strings.Trim(string(samlSignedRequest), "\n")
	return samlSignedRequestXML, nil
}

// VerifyResponseSignature verify signature of a SAML 2.0 Response document
func VerifyResponseSignature(xml string, publicCert string) error {
	return verify(xml, publicCert, xmlResponseID)
}

// VerifyResponseSignature verify signature of a SAML 2.0 Assertion document
func VerifyAssertionSignature(xml string, publicCert string) error {
	return verify(xml, publicCert, xmlAssertionID)
}

// VerifyRequestSignature verify signature of a SAML 2.0 AuthnRequest document
func VerifyRequestSignature(xml string, publicCert string) error {
	return verify(xml, publicCert, xmlRequestID)
}

func verify(xml string, publicCert string, id string) error {

	publicCertFile, err := writeToTemp(publicCert)
	if err != nil {
		return err
	}
	defer deleteTempFile(publicCertFile.Name())

	samlXmlsecInput, err := writeToTemp(xml)
	if err != nil {
		return err
	}
	defer deleteTempFile(samlXmlsecInput.Name())

	output, err := exec.Command("xmlsec1", "--verify", "--pubkey-cert-pem", publicCertFile.Name(), "--id-attr:ID", id, samlXmlsecInput.Name()).CombinedOutput()
	if err != nil {
		return errors.New(err.Error() + " : " + string(output))
	}
	return nil
}

// DefaultSignature returns a Signature struct that uses the default c14n and SHA1 settings.
func DefaultSignature(pemEncodedPublicKey string) Signature {
	// xmlsec wants the key to be base64-encoded but *not* wrapped with the
	// PEM flags
	pemBlock, _ := pem.Decode([]byte(pemEncodedPublicKey))
	certStr := base64.StdEncoding.EncodeToString(pemBlock.Bytes)

	return Signature{
		Id: "Signature1",
		SignedInfo: SignedInfo{
			CanonicalizationMethod: Method{
				Algorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
			},
			SignatureMethod: Method{
				Algorithm: "http://www.w3.org/2000/09/xmldsig#rsa-sha1",
			},
			Reference: Reference{
				ReferenceTransforms: []Method{
					Method{Algorithm: "http://www.w3.org/2000/09/xmldsig#enveloped-signature"},
				},
				DigestMethod: Method{
					Algorithm: "http://www.w3.org/2000/09/xmldsig#sha1",
				},
			},
		},
		X509Certificate: &SignatureX509Data{
			X509Certificate: certStr,
		},
	}
}

const encTempl = `<EncryptedData Type="http://www.w3.org/2001/04/xmlenc#Element" xmlns="http://www.w3.org/2001/04/xmlenc#">
	<EncryptionMethod Algorithm="http://www.w3.org/2001/04/xmlenc#aes256-cbc"/>
	<KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
		<EncryptedKey xmlns="http://www.w3.org/2001/04/xmlenc#">
			<EncryptionMethod Algorithm="http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p">
				<DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1" xmlns="http://www.w3.org/2000/09/xmldsig#"/>
			</EncryptionMethod>
			<KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
				<X509Data>
					<X509Certificate/>
				</X509Data>
			</KeyInfo>
			<CipherData>
				<CipherValue/>
			</CipherData>
		</EncryptedKey>
	</KeyInfo>
	<CipherData>
		<CipherValue/>
	</CipherData>
</EncryptedData>`

// Encrypt encrypt an xml plaintext value with a private key
func Encrypt(plaintext string, publicKey string) (string, error) {

	publicKeyFile, err := writeToTemp(publicKey)
	if err != nil {
		return "", err
	}
	defer deleteTempFile(publicKeyFile.Name())

	samlXmlsecInput, err := writeToTemp(plaintext)
	if err != nil {
		return "", err
	}
	defer deleteTempFile(samlXmlsecInput.Name())

	samlXmlsecTemplate, err := writeToTemp(encTempl)
	if err != nil {
		return "", err
	}
	defer deleteTempFile(samlXmlsecTemplate.Name())

	samlXmlsecOutput, err := ioutil.TempFile(os.TempDir(), "tmpgs")
	if err != nil {
		return "", err
	}
	defer deleteTempFile(samlXmlsecOutput.Name())
	samlXmlsecOutput.Close()

	output, err := exec.Command("xmlsec1", "--encrypt", "--session-key", "aes-256-cbc", "--pubkey-cert-pem", publicKeyFile.Name(),
		"--output", samlXmlsecOutput.Name(), "--xml-data", samlXmlsecInput.Name(), samlXmlsecTemplate.Name()).CombinedOutput()
	if err != nil {
		return "", errors.New(err.Error() + " : " + string(output))
	}

	encrypted, err := ioutil.ReadFile(samlXmlsecOutput.Name())
	if err != nil {
		return "", err
	}
	encryptedXML := strings.Trim(string(encrypted), "\n")
	return encryptedXML, nil
}

// Decrypt decrypt an xml cipher value with a third-party public key
func Decrypt(cipher string, privateKey string) (string, error) {

	privateKeyFile, err := writeToTemp(privateKey)
	if err != nil {
		return "", err
	}
	defer deleteTempFile(privateKeyFile.Name())

	samlXmlsecInput, err := writeToTemp(cipher)
	if err != nil {
		return "", err
	}
	defer deleteTempFile(samlXmlsecInput.Name())

	samlXmlsecOutput, err := ioutil.TempFile(os.TempDir(), "tmpgs")
	if err != nil {
		return "", err
	}
	defer deleteTempFile(samlXmlsecOutput.Name())
	samlXmlsecOutput.Close()

	output, err := exec.Command("xmlsec1", "--decrypt", "--privkey-pem", privateKeyFile.Name(), "--id-attr:ID", "http://www.w3.org/2001/04/xmlenc#EncryptedData",
		"--output", samlXmlsecOutput.Name(), samlXmlsecInput.Name()).CombinedOutput()
	if err != nil {
		return "", errors.New(err.Error() + " : " + string(output))
	}

	decrypted, err := ioutil.ReadFile(samlXmlsecOutput.Name())
	if err != nil {
		return "", err
	}
	decryptedXML := strings.Trim(string(decrypted), "\n")
	return decryptedXML, nil
}

// writeToTemp write a string to a temporary file and close
// Caller is responsible for its lifetime
func writeToTemp(val string) (*os.File, error) {
	f, err := ioutil.TempFile(os.TempDir(), "tf")
	if err != nil {
		return nil, err
	}

	f.WriteString(val)
	f.Close()

	return f, nil
}

// deleteTempFile remove a file and ignore error
// Intended to be called in a defer after the creation of a temp file to ensure cleanup
func deleteTempFile(filename string) {
	_ = os.Remove(filename)
}
