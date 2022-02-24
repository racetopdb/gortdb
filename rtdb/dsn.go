package rtdb

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultCharset = "iso-8859-1"
	defaultAddr    = "127.0.0.1:9000"
	defaultHost    = "127.0.0.1"
	defaultPort    = 9000
)

type Config struct {
	prepareConfig
	User         string            // Username
	Password     string            // Password
	Protocol     string            // Net protocol type
	Address      string            // Network address
	DBName       string            // Database name
	Location     *time.Location    // Time zone setting
	DialTimeout  time.Duration     // Dial timeout
	ReadTimeout  time.Duration     // I/O read timeout
	WriteTimeout time.Duration     // I/O write timeout
	Charset      string            // Character set
	Params       map[string]string // Connection parameters
	ParseTime    bool              // Parse time values to time.Time
}

func NewConfig() *Config {
	c := &Config{
		Protocol:    "tcp",
		Address:     "127.0.0.1:9000",
		DialTimeout: time.Millisecond * 500,
		Charset:     defaultCharset,
		Location:    time.UTC,
	}

	return c
}

func (c *Config) HostAndPort() (string, int) {
	if c.Address != "" {
		parts := strings.Split(c.Address, ":")
		if len(parts) != 2 {
			return defaultHost, defaultPort
		}
		port, _ := strconv.Atoi(parts[1])
		return parts[0], port
	}
	return defaultHost, defaultPort
}

// adjust represent adjust Config's field value from dsn params.
func (c *Config) adjust() error {
	params := c.Params
	if len(params) == 0 {
		return nil
	}
	var charset, loc, parseTime string
	for k, v := range params {
		switch k {
		case "charset":
			charset = v
		case "parseTime":
			parseTime = v
		case "loc":
			loc = v
		default:
			// no option to adjust
			return nil
		}
	}
	if err := c.prepare(loc, charset, parseTime); err != nil {
		return err
	}

	return nil
}

func (c *Config) prepare(loc, charset, parseTime string) error {
	if loc != "" {
		if err := c.parseLoc(loc); err != nil {
			return err
		}
		c.ParseTime = c._parseTime
	}
	if charset != "" {
		if err := c.parseCharset(charset); err != nil {
			return err
		}
		c.Charset = c._charset
	}
	if parseTime != "" {
		if err := c.parseTime(parseTime); err != nil {
			return err
		}
		c.ParseTime = c._parseTime
	}
	c.prepareConfig.revert()

	return nil
}

type prepareConfig struct {
	_loc       *time.Location
	_charset   string
	_parseTime bool
}

// revert just for test functions.
func (pc *prepareConfig) revert() {
	pc._loc = nil
	pc._charset = ""
	pc._parseTime = false
}

func (pc *prepareConfig) parseLoc(loc string) error {
	v, err := url.QueryUnescape(loc)
	if err != nil {
		return err
	}
	loctmp, err := time.LoadLocation(v)
	if err != nil {
		return err
	}
	pc._loc = loctmp
	return nil
}

func (pc *prepareConfig) parseCharset(charset string) error {
	charset = strings.ToLower(charset)
	charsetId, ok := charsetMap[charset]
	if !ok {
		rtdbLogger.Printf("parse charset failed\n")
		return InvalidDSN
	}
	if charsetId != CHARSET_UNKNOWN {
		pc._charset = charset
	}
	return nil
}

func (pc *prepareConfig) parseTime(parseTime string) error {
	switch parseTime {
	case "true", "t", "T", "True", "1":
		pc._parseTime = true
	case "false", "f", "F", "False", "0":
		pc._parseTime = false
	default:
		rtdbLogger.Printf("parse time failed\n")
		return InvalidDSN
	}
	return nil
}

// ParseDSN parses the dsn+ to a Config.
// The DSN format is "{user}:{password}@{protocol}({host}:{port})/{dbname}?{param1}={value1}&{param2}={value2}"
func ParseDSN(dsn string) (config *Config, err error) {
	config = NewConfig()

	foundLastSlash := false

	// TODO: fix time zone bug
	// var boudaryIndex int
	// boudaryIndex = strings.LastIndex(dsn, "/")
	// locIndex := strings.LastIndex(dsn, "loc")
	// if boudaryIndex == -1 && len(dsn) > 0 {
	// 	return nil, InvalidDSN
	// }
	// if boudaryIndex > 0 {
	// 	if boudaryIndex > locIndex {
	// 		lastSlashIndex := strings.LastIndex(dsn[:boudaryIndex], "/")
	// 		if lastSlashIndex == -1 && len(dsn) > 0 {
	// 			return nil, InvalidDSN
	// 		}
	// 		if lastSlashIndex == 0 {
	// 			questionMarkIndex := strings.IndexByte(dsn, '?')
	// 			// not found params
	// 			if questionMarkIndex == -1 {
	// 				config.DBName = dsn[1:]
	// 			} else {
	// 				config.DBName = dsn[1:questionMarkIndex]
	// 				config.Params = parseDSNParams(dsn[questionMarkIndex+1:])
	// 			}
	// 		}
	// 	}
	// }
	for i := len(dsn) - 1; i >= 0; i-- {
		// how to handle time zone?? example .....Asia/Shanghai....
		if dsn[i] == '/' {
			foundLastSlash = true
			if i > 0 {
				lastSlashIndex := i
				// deal with left part
				var userInfoRightIndex int
				// find the last '@'
				for j := i - 1; j >= 0; j-- {
					// find the host and port
					if dsn[j] == '@' {
						var lbracketsIndex, rbracketsIndex int
						for k := lastSlashIndex - 1; k > j; k-- {
							if dsn[k] == ')' {
								rbracketsIndex = k
							}
							if dsn[k] == '(' {
								lbracketsIndex = k
							}
						}
						if lbracketsIndex >= 0 && rbracketsIndex > 0 {
							config.Address = dsn[lbracketsIndex+1 : rbracketsIndex]
							config.Protocol = dsn[j+1 : lbracketsIndex]
						} else {
							config.Protocol = dsn[j+1 : lastSlashIndex]
						}
						userInfoRightIndex = j
						break
					}
				}
				// find the first ':'
				if userInfoRightIndex == 0 {
					userInfoRightIndex = i
				}
				colonIndex := strings.IndexByte(dsn[:userInfoRightIndex], ':')
				if colonIndex == -1 {
					config.User = dsn[:userInfoRightIndex]
				} else {
					config.User = dsn[:colonIndex]
					config.Password = dsn[colonIndex+1 : userInfoRightIndex]
				}

				// deal with the right part
				// bug!!!这将会找到第一个问号而不是最后一个问号
				// it will find question mark in the / right range
				right := dsn[lastSlashIndex+1:]
				questionMarkIndex := strings.IndexByte(right, '?')
				if questionMarkIndex == -1 {
					// not found params
					config.DBName = dsn[lastSlashIndex+1:]
				} else {
					config.DBName = right[:questionMarkIndex]
					config.Params = parseDSNParams(right[questionMarkIndex+1:])
				}
			} else {
				// found '/' at zero index
				questionMarkIndex := strings.IndexByte(dsn, '?')
				// not found params
				if questionMarkIndex == -1 {
					config.DBName = dsn[1:]
				} else {
					config.DBName = dsn[1:questionMarkIndex]
					config.Params = parseDSNParams(dsn[questionMarkIndex+1:])
				}
			}
			break
		}
	}
	if !foundLastSlash && len(dsn) > 0 {
		rtdbLogger.Printf("slash not found\n")
		return nil, InvalidDSN
	}
	if err := config.adjust(); err != nil {
		return nil, err
	}
	return config, nil
}

func parseDSNParams(params string) map[string]string {
	kv := make(map[string]string)
	kvSlice := strings.Split(params, "&")
	if len(kvSlice) == 0 {
		return nil
	}

	for _, kvStr := range kvSlice {
		kvTmp := strings.Split(kvStr, "=")
		if len(kvTmp) != 2 {
			rtdbLogger.Printf("parse kv pair failed\n")
			continue
		}
		kv[kvTmp[0]] = kvTmp[1]
	}

	return kv
}

const (
	CHARSET_UNKNOWN int8 = iota
	CHARSET_GBK
	CHARSET_UTF8
	CHARSET_UCS2LE
	CHARSET_UCS2BE
	CHARSET_BIG5
	CHARSET_EUCJP
	CHARSET_SJIS
	CHARSET_EUCKR
	CHARSET_ISO1
	CHARSET_WIN1
	CHARSET_WIN2
)

var charsetMap = map[string]int8{
	"":             CHARSET_UNKNOWN,
	"gbk":          CHARSET_GBK,
	"utf-8":        CHARSET_UTF8,
	"ucs-2le":      CHARSET_UCS2LE,
	"ucs-2be":      CHARSET_UCS2BE,
	"big-5":        CHARSET_BIG5,
	"euc-jp":       CHARSET_EUCJP,
	"shift-jis":    CHARSET_SJIS,
	"euc-kr":       CHARSET_EUCKR,
	"iso-8859-1":   CHARSET_ISO1,
	"windows-1251": CHARSET_WIN1,
	"windows-1252": CHARSET_WIN2,
}
