package strfmt

import (
	validator "github.com/octohelm/courier/pkg/validator"
)

func init() { validator.Register(ASCIIValidator) }

var ASCIIValidator = validator.NewRegexpStrfmtValidator(regexpStringASCII, "ascii")

func init() { validator.Register(AlphaValidator) }

var AlphaValidator = validator.NewRegexpStrfmtValidator(regexpStringAlpha, "alpha")

func init() { validator.Register(AlphaNumericValidator) }

var AlphaNumericValidator = validator.NewRegexpStrfmtValidator(regexpStringAlphaNumeric, "alpha-numeric", "alphaNumeric")

func init() { validator.Register(AlphaUnicodeValidator) }

var AlphaUnicodeValidator = validator.NewRegexpStrfmtValidator(regexpStringAlphaUnicode, "alpha-unicode", "alphaUnicode")

func init() {
	validator.Register(AlphaUnicodeNumericValidator)
}

var AlphaUnicodeNumericValidator = validator.NewRegexpStrfmtValidator(regexpStringAlphaUnicodeNumeric, "alpha-unicode-numeric", "alphaUnicodeNumeric")

func init() { validator.Register(Base64Validator) }

var Base64Validator = validator.NewRegexpStrfmtValidator(regexpStringBase64, "base64")

func init() { validator.Register(Base64URLValidator) }

var Base64URLValidator = validator.NewRegexpStrfmtValidator(regexpStringBase64URL, "base64-url", "base64URL")

func init() { validator.Register(BtcAddressValidator) }

var BtcAddressValidator = validator.NewRegexpStrfmtValidator(regexpStringBtcAddress, "btc-address", "btcAddress")

func init() { validator.Register(BtcAddressLowerValidator) }

var BtcAddressLowerValidator = validator.NewRegexpStrfmtValidator(regexpStringBtcAddressLower, "btc-address-lower", "btcAddressLower")

func init() { validator.Register(BtcAddressUpperValidator) }

var BtcAddressUpperValidator = validator.NewRegexpStrfmtValidator(regexpStringBtcAddressUpper, "btc-address-upper", "btcAddressUpper")

func init() { validator.Register(DataURIValidator) }

var DataURIValidator = validator.NewRegexpStrfmtValidator(regexpStringDataURI, "data-uri", "dataURI")

func init() { validator.Register(EmailValidator) }

var EmailValidator = validator.NewRegexpStrfmtValidator(regexpStringEmail, "email")

func init() { validator.Register(EthAddressValidator) }

var EthAddressValidator = validator.NewRegexpStrfmtValidator(regexpStringEthAddress, "eth-address", "ethAddress")

func init() { validator.Register(EthAddressLowerValidator) }

var EthAddressLowerValidator = validator.NewRegexpStrfmtValidator(regexpStringEthAddressLower, "eth-address-lower", "ethAddressLower")

func init() { validator.Register(EthAddressUpperValidator) }

var EthAddressUpperValidator = validator.NewRegexpStrfmtValidator(regexpStringEthAddressUpper, "eth-address-upper", "ethAddressUpper")

func init() { validator.Register(HslValidator) }

var HslValidator = validator.NewRegexpStrfmtValidator(regexpStringHSL, "hsl")

func init() { validator.Register(HslaValidator) }

var HslaValidator = validator.NewRegexpStrfmtValidator(regexpStringHSLA, "hsla")

func init() { validator.Register(HexAdecimalValidator) }

var HexAdecimalValidator = validator.NewRegexpStrfmtValidator(regexpStringHexAdecimal, "hex-adecimal", "hexAdecimal")

func init() { validator.Register(HexColorValidator) }

var HexColorValidator = validator.NewRegexpStrfmtValidator(regexpStringHexColor, "hex-color", "hexColor")

func init() { validator.Register(HostnameValidator) }

var HostnameValidator = validator.NewRegexpStrfmtValidator(regexpStringHostname, "hostname")

func init() { validator.Register(HostnameXValidator) }

var HostnameXValidator = validator.NewRegexpStrfmtValidator(regexpStringHostnameX, "hostname-x", "hostnameX")

func init() { validator.Register(Isbn10Validator) }

var Isbn10Validator = validator.NewRegexpStrfmtValidator(regexpStringISBN10, "isbn10")

func init() { validator.Register(Isbn13Validator) }

var Isbn13Validator = validator.NewRegexpStrfmtValidator(regexpStringISBN13, "isbn13")

func init() { validator.Register(LatitudeValidator) }

var LatitudeValidator = validator.NewRegexpStrfmtValidator(regexpStringLatitude, "latitude")

func init() { validator.Register(LongitudeValidator) }

var LongitudeValidator = validator.NewRegexpStrfmtValidator(regexpStringLongitude, "longitude")

func init() { validator.Register(MultibyteValidator) }

var MultibyteValidator = validator.NewRegexpStrfmtValidator(regexpStringMultibyte, "multibyte")

func init() { validator.Register(NumberValidator) }

var NumberValidator = validator.NewRegexpStrfmtValidator(regexpStringNumber, "number")

func init() { validator.Register(NumericValidator) }

var NumericValidator = validator.NewRegexpStrfmtValidator(regexpStringNumeric, "numeric")

func init() { validator.Register(PrintableASCIIValidator) }

var PrintableASCIIValidator = validator.NewRegexpStrfmtValidator(regexpStringPrintableASCII, "printable-ascii", "printableASCII")

func init() { validator.Register(RgbValidator) }

var RgbValidator = validator.NewRegexpStrfmtValidator(regexpStringRGB, "rgb")

func init() { validator.Register(RgbaValidator) }

var RgbaValidator = validator.NewRegexpStrfmtValidator(regexpStringRGBA, "rgba")

func init() { validator.Register(SsnValidator) }

var SsnValidator = validator.NewRegexpStrfmtValidator(regexpStringSSN, "ssn")

func init() { validator.Register(UUIDValidator) }

var UUIDValidator = validator.NewRegexpStrfmtValidator(regexpStringUUID, "uuid")

func init() { validator.Register(Uuid3Validator) }

var Uuid3Validator = validator.NewRegexpStrfmtValidator(regexpStringUUID3, "uuid3")

func init() { validator.Register(Uuid4Validator) }

var Uuid4Validator = validator.NewRegexpStrfmtValidator(regexpStringUUID4, "uuid4")

func init() { validator.Register(Uuid5Validator) }

var Uuid5Validator = validator.NewRegexpStrfmtValidator(regexpStringUUID5, "uuid5")
