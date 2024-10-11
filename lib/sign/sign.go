package sign

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	TokenLength = 107
	BaseStr     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789=_"
	Url         = "https://live.douyin.com/"
	UserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

const (
	FailedToCreateRequestError = "FailedToCreateRequestError: %v"
	FailedToParseWssUrlError   = "FailedToParseWssUrlError: %v"
	FailedToCallSignError      = "FailedToCallSignError: %v"
)

func GenerateMsToken() string {
	baseLen := len(BaseStr)
	rand.New(rand.NewSource(time.Now().UnixNano()))

	var randomStr strings.Builder
	for i := 0; i < TokenLength; i++ {
		randomStr.WriteByte(BaseStr[rand.Intn(baseLen)])
	}

	return randomStr.String()
}

func GenerateSignature(wss string) (string, error) {
	params := []string{"live_id", "aid", "version_code", "webcast_sdk_version",
		"room_id", "sub_room_id", "sub_channel_id", "did_rule",
		"user_unique_id", "device_platform", "device_type", "ac",
		"identity"}

	u, err := url.Parse(wss)
	if err != nil {
		return "", fmt.Errorf(FailedToParseWssUrlError, err)
	}
	queryParams := u.Query()

	var tplParams []string
	for _, param := range params {
		value := queryParams.Get(param)
		tplParams = append(tplParams, fmt.Sprintf("%s=%s", param, value))
	}

	paramStr := strings.Join(tplParams, ",")
	hash := md5.New()
	hash.Write([]byte(paramStr))
	md5Param := hex.EncodeToString(hash.Sum(nil))

	scriptFile := "./lib/js/sign.js"

	param := map[string]string{"X-MS-STUB": md5Param}
	jsonParams, _ := json.Marshal(param)
	cmd := exec.Command("node", scriptFile, string(jsonParams))
	cmd.Env = append(os.Environ(), "LANG=en_US.UTF-8")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf(FailedToCallSignError, err)
	}

	var result map[string]string
	_ = json.Unmarshal(out.Bytes(), &result)
	return result["X-Bogus"], nil
}
