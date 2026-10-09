package config

import (
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var ValueOf = &config{}

type allowedUsers []int64

func (au *allowedUsers) Decode(value string) error {
	if value == "" {
		return nil
	}
	ids := strings.Split(string(value), ",")
	for _, id := range ids {
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return err
		}
		*au = append(*au, idInt)
	}
	return nil
}

type config struct {
	ApiID          int32        `envconfig:"API_ID" required:"true"`
	ApiHash        string       `envconfig:"API_HASH" required:"true"`
	BotToken       string       `envconfig:"BOT_TOKEN" required:"true"`
	LogChannelID   int64        `envconfig:"LOG_CHANNEL" required:"true"`
	Dev            bool         `envconfig:"DEV" default:"false"`
	Port           int          `envconfig:"PORT" default:"8080"`
	Host           string       `envconfig:"HOST" default:""`
	HashLength     int          `envconfig:"HASH_LENGTH" default:"6"`
	UseSessionFile bool         `envconfig:"USE_SESSION_FILE" default:"true"`
	UserSession    string       `envconfig:"USER_SESSION"`
	UsePublicIP    bool         `envconfig:"USE_PUBLIC_IP" default:"true"`
	AllowedUsers   allowedUsers `envconfig:"ALLOWED_USERS"`
	MultiTokens    []string

	// stream specific config
	StreamConcurrency int `envconfig:"STREAM_CONCURRENCY" default:"4"`
	StreamBufferCount int `envconfig:"STREAM_BUFFER_COUNT" default:"8"`
	StreamTimeoutSec  int `envconfig:"STREAM_TIMEOUT_SEC" default:"30"`
	StreamMaxRetries  int `envconfig:"STREAM_MAX_RETRIES" default:"3"`
}

var botTokenRegex = regexp.MustCompile(`MULTI\_TOKEN\d+=(.*)`)

func (c *config) loadFromEnvFile(log *zap.Logger) {
	envPath := filepath.Clean("fsb.env")
	log.Sugar().Infof("Trying to load ENV vars from %s", envPath)
	err := godotenv.Load(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Sugar().Errorf("ENV file not found: %s", envPath)
		} else {
			log.Fatal("Unknown error while parsing env file.", zap.Error(err))
		}
	}
}

func SetFlagsFromConfig(cmd *cobra.Command) {
	cmd.Flags().Int32("api-id", ValueOf.ApiID, "Telegram API ID")
	cmd.Flags().String("api-hash", ValueOf.ApiHash, "Telegram API Hash")
	cmd.Flags().String("bot-token", ValueOf.BotToken, "Telegram Bot Token")
	cmd.Flags().Int64("log-channel", ValueOf.LogChannelID, "Telegram Log Channel ID")
	cmd.Flags().Bool("dev", ValueOf.Dev, "Enable development mode")
	cmd.Flags().IntP("port", "p", ValueOf.Port, "Server port")
	cmd.Flags().String("host", ValueOf.Host, "Server host that will be included in links")
	cmd.Flags().Int("hash-length", ValueOf.HashLength, "Hash length in links")
	cmd.Flags().Bool("use-session-file", ValueOf.UseSessionFile, "Use session files")
	cmd.Flags().String("user-session", ValueOf.UserSession, "Pyrogram user session")
	cmd.Flags().Bool("use-public-ip", ValueOf.UsePublicIP, "Use public IP instead of local IP")
	cmd.Flags().String("multi-token-txt-file", "", "Multi token txt file (Not implemented)")
	cmd.Flags().Int("stream-concurrency", ValueOf.StreamConcurrency, "Number of parallel block fetches")
	cmd.Flags().Int("stream-buffer-count", ValueOf.StreamBufferCount, "Number of blocks to prefetch")
	cmd.Flags().Int("stream-timeout-sec", ValueOf.StreamTimeoutSec, "Maximum time to wait for a single block (in seconds)")
	cmd.Flags().Int("stream-max-retries", ValueOf.StreamMaxRetries, "Number of retry attempts for failed fetches")
}

func (c *config) loadConfigFromArgs(log *zap.Logger, cmd *cobra.Command) {
	apiID, _ := cmd.Flags().GetInt32("api-id")
	if apiID != 0 {
		os.Setenv("API_ID", strconv.Itoa(int(apiID)))
	}
	apiHash, _ := cmd.Flags().GetString("api-hash")
	if apiHash != "" {
		os.Setenv("API_HASH", apiHash)
	}
	botToken, _ := cmd.Flags().GetString("bot-token")
	if botToken != "" {
		os.Setenv("BOT_TOKEN", botToken)
	}
	logChannelID, _ := cmd.Flags().GetInt64("log-channel")
	if logChannelID != 0 {
		os.Setenv("LOG_CHANNEL", strconv.FormatInt(logChannelID, 10))
	}
	dev, _ := cmd.Flags().GetBool("dev")
	if dev {
		os.Setenv("DEV", strconv.FormatBool(dev))
	}
	port, _ := cmd.Flags().GetInt("port")
	if port != 0 {
		os.Setenv("PORT", strconv.Itoa(port))
	}
	host, _ := cmd.Flags().GetString("host")
	if host != "" {
		os.Setenv("HOST", host)
	}
	hashLength, _ := cmd.Flags().GetInt("hash-length")
	if hashLength != 0 {
		os.Setenv("HASH_LENGTH", strconv.Itoa(hashLength))
	}
	useSessionFile, _ := cmd.Flags().GetBool("use-session-file")
	if useSessionFile {
		os.Setenv("USE_SESSION_FILE", strconv.FormatBool(useSessionFile))
	}
	userSession, _ := cmd.Flags().GetString("user-session")
	if userSession != "" {
		os.Setenv("USER_SESSION", userSession)
	}
	usePublicIP, _ := cmd.Flags().GetBool("use-public-ip")
	if usePublicIP {
		os.Setenv("USE_PUBLIC_IP", strconv.FormatBool(usePublicIP))
	}
	multiTokens, _ := cmd.Flags().GetString("multi-token-txt-file")
	if multiTokens != "" {
		os.Setenv("MULTI_TOKEN_TXT_FILE", multiTokens)
	}
	streamConcurrency, _ := cmd.Flags().GetInt("stream-concurrency")
	if streamConcurrency != 0 {
		os.Setenv("STREAM_CONCURRENCY", strconv.Itoa(streamConcurrency))
	}
	streamBufferCount, _ := cmd.Flags().GetInt("stream-buffer-count")
	if streamBufferCount != 0 {
		os.Setenv("STREAM_BUFFER_COUNT", strconv.Itoa(streamBufferCount))
	}
	streamTimeoutSec, _ := cmd.Flags().GetInt("stream-timeout-sec")
	if streamTimeoutSec != 0 {
		os.Setenv("STREAM_TIMEOUT_SEC", strconv.Itoa(streamTimeoutSec))
	}
	streamMaxRetries, _ := cmd.Flags().GetInt("stream-max-retries")
	if streamMaxRetries != 0 {
		os.Setenv("STREAM_MAX_RETRIES", strconv.Itoa(streamMaxRetries))
	}
}

func (c *config) setupEnvVars(log *zap.Logger, cmd *cobra.Command) {
	c.loadFromEnvFile(log)
	c.loadConfigFromArgs(log, cmd)
	err := envconfig.Process("", c)
	if err != nil {
		log.Fatal("Error while parsing env variables", zap.Error(err))
	}

	ip, err := getIP(c.UsePublicIP)
	if err != nil {
		log.Error("Error while getting IP", zap.Error(err))
	}

	if c.Host == "" {
		c.Host = "http://" + ip + ":" + strconv.Itoa(c.Port)
		log.Sugar().Info("HOST set automatically to " + c.Host)
	}

	val := reflect.ValueOf(c).Elem()
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "MULTI_TOKEN") {
			match := botTokenRegex.FindStringSubmatch(env)
			if len(match) > 1 {
				c.MultiTokens = append(c.MultiTokens, match[1])
			}
		}
	}
	val.FieldByName("MultiTokens").Set(reflect.ValueOf(c.MultiTokens))
}

func Load(log *zap.Logger, cmd *cobra.Command) {
	log = log.Named("Config")
	defer log.Info("Loaded config")
	ValueOf.setupEnvVars(log, cmd)
	ValueOf.LogChannelID = int64(stripInt(log, int(ValueOf.LogChannelID)))
	if ValueOf.HashLength == 0 {
		ValueOf.HashLength = 6
	}
}

func getIP(public bool) (string, error) {
	var ip string
	var err error
	if public {
		ip, err = GetPublicIP()
	} else {
		ip, err = getInternalIP()
	}
	if ip == "" {
		ip = "localhost"
	}
	if err != nil {
		return "localhost", err
	}
	return ip, nil
}

func getInternalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", errors.New("no internet connection")
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(ip), nil
}

func stripInt(log *zap.Logger, a int) int {
	strA := strconv.Itoa(abs(a))
	lastDigits := strings.Replace(strA, "100", "", 1)
	result, err := strconv.Atoi(lastDigits)
	if err != nil {
		log.Sugar().Fatalln(err)
		return 0
	}
	return result
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
