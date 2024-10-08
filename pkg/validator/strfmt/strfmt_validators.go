package strfmt

import (
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/validators"
)

func init() { internal.Register(ASCIIValidatorProvider) }

var ASCIIValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringASCII, "ascii")

func init() { internal.Register(AlphaValidatorProvider) }

var AlphaValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringAlpha, "alpha")

func init() { internal.Register(AlphaNumericValidatorProvider) }

var AlphaNumericValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringAlphaNumeric, "alpha-numeric", "alphaNumeric")

func init() { internal.Register(AlphaUnicodeValidatorProvider) }

var AlphaUnicodeValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringAlphaUnicode, "alpha-unicode", "alphaUnicode")

func init() {
	internal.Register(AlphaUnicodeNumericValidatorProvider)
}

var AlphaUnicodeNumericValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringAlphaUnicodeNumeric, "alpha-unicode-numeric", "alphaUnicodeNumeric")

func init() { internal.Register(Base64ValidatorProvider) }

var Base64ValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringBase64, "base64")

func init() { internal.Register(Base64URLValidatorProvider) }

var Base64URLValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringBase64URL, "base64-url", "base64URL")

func init() { internal.Register(BtcAddressValidatorProvider) }

var BtcAddressValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringBtcAddress, "btc-address", "btcAddress")

func init() { internal.Register(BtcAddressLowerValidatorProvider) }

var BtcAddressLowerValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringBtcAddressLower, "btc-address-lower", "btcAddressLower")

func init() { internal.Register(BtcAddressUpperValidatorProvider) }

var BtcAddressUpperValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringBtcAddressUpper, "btc-address-upper", "btcAddressUpper")

func init() { internal.Register(DataURIValidatorProvider) }

var DataURIValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringDataURI, "data-uri", "dataURI")

func init() { internal.Register(EmailValidatorProvider) }

var EmailValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringEmail, "email")

func init() { internal.Register(EthAddressValidatorProvider) }

var EthAddressValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringEthAddress, "eth-address", "ethAddress")

func init() { internal.Register(EthAddressLowerValidatorProvider) }

var EthAddressLowerValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringEthAddressLower, "eth-address-lower", "ethAddressLower")

func init() { internal.Register(EthAddressUpperValidatorProvider) }

var EthAddressUpperValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringEthAddressUpper, "eth-address-upper", "ethAddressUpper")

func init() { internal.Register(HslValidatorProvider) }

var HslValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringHSL, "hsl")

func init() { internal.Register(HslaValidatorProvider) }

var HslaValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringHSLA, "hsla")

func init() { internal.Register(HexAdecimalValidatorProvider) }

var HexAdecimalValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringHexAdecimal, "hex-adecimal", "hexAdecimal")

func init() { internal.Register(HexColorValidatorProvider) }

var HexColorValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringHexColor, "hex-color", "hexColor")

func init() { internal.Register(HostnameValidatorProvider) }

var HostnameValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringHostname, "hostname")

func init() { internal.Register(HostnameXValidatorProvider) }

var HostnameXValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringHostnameX, "hostname-x", "hostnameX")

func init() { internal.Register(Isbn10ValidatorProvider) }

var Isbn10ValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringISBN10, "isbn10")

func init() { internal.Register(Isbn13ValidatorProvider) }

var Isbn13ValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringISBN13, "isbn13")

func init() { internal.Register(LatitudeValidatorProvider) }

var LatitudeValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringLatitude, "latitude")

func init() { internal.Register(LongitudeValidatorProvider) }

var LongitudeValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringLongitude, "longitude")

func init() { internal.Register(MultibyteValidatorProvider) }

var MultibyteValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringMultibyte, "multibyte")

func init() { internal.Register(NumberValidatorProvider) }

var NumberValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringNumber, "number")

func init() { internal.Register(NumericValidatorProvider) }

var NumericValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringNumeric, "numeric")

func init() { internal.Register(PrintableASCIIValidatorProvider) }

var PrintableASCIIValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringPrintableASCII, "printable-ascii", "printableASCII")

func init() { internal.Register(RgbValidatorProvider) }

var RgbValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringRGB, "rgb")

func init() { internal.Register(RgbaValidatorProvider) }

var RgbaValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringRGBA, "rgba")

func init() { internal.Register(SsnValidatorProvider) }

var SsnValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringSSN, "ssn")

func init() { internal.Register(UUIDValidatorProvider) }

var UUIDValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringUUID, "uuid")

func init() { internal.Register(Uuid3ValidatorProvider) }

var Uuid3ValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringUUID3, "uuid3")

func init() { internal.Register(Uuid4ValidatorProvider) }

var Uuid4ValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringUUID4, "uuid4")

func init() { internal.Register(Uuid5ValidatorProvider) }

var Uuid5ValidatorProvider = validators.NewRegexpStrfmtValidatorProvider(regexpStringUUID5, "uuid5")
