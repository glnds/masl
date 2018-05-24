package xmlsec

import (
	"encoding/xml"
)

// Method is part of Signature.
type Method struct {
	Algorithm string `xml:",attr"`
}

// Signature is a model for the Signature object specified by XMLDSIG. This is
// convenience object when constructing XML that you'd like to sign. For example:
//
//    type Foo struct {
//       Stuff string
//       Signature Signature
//    }
//
//    f := Foo{Suff: "hello"}
//    f.Signature = DefaultSignature()
//    buf, _ := xml.Marshal(f)
//    buf, _ = Sign(key, buf)
//
type Signature struct {
	XMLName xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# Signature"`

	Id              string             `xml:"Id,attr"`
	SignedInfo      SignedInfo         `xml:"SignedInfo"`
	SignatureValue  string             `xml:"SignatureValue"`
	KeyName         string             `xml:"KeyInfo>KeyName,omitempty"`
	X509Certificate *SignatureX509Data `xml:"KeyInfo>X509Data,omitempty"`
}

type SignedInfo struct {
	CanonicalizationMethod Method    `xml:"CanonicalizationMethod"`
	SignatureMethod        Method    `xml:"SignatureMethod"`
	Reference              Reference `xml:"Reference"`
}

type Reference struct {
	URI                 string   `xml:"URI,attr"`
	ReferenceTransforms []Method `xml:"Transforms>Transform"`
	DigestMethod        Method   `xml:"DigestMethod"`
	DigestValue         string   `xml:"DigestValue"`
}

// SignatureX509Data represents the <X509Data> element of <Signature>
type SignatureX509Data struct {
	X509Certificate string `xml:"X509Certificate,omitempty"`
}
