package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"service-outpatient/config"
	"service-outpatient/datastruct"
	"service-outpatient/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Getter struct {
	NoIHS       string
	RefID       string
	ServiceName datastruct.ServiceName
}

func PostRequest(c *gin.Context, serviceName datastruct.ServiceName, body []byte) (string, error) {
	bufferBytes := bytes.NewBuffer(body)

	var serviceUrl string
	switch serviceName {
	case "laboratory":
		serviceUrl = config.LabServiceURL
	case "radiology":
		serviceUrl = config.RadiologyServiceURL
	case "pharmacy":
		serviceUrl = config.PharmacyServiceURL
	default:
		return "", fmt.Errorf("service %s undefined", serviceName)
	}

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/v1/request/%s", serviceUrl, serviceName),
		bufferBytes,
	)
	req.Header.Add("Authorization", c.GetHeader("Authorization"))
	req.Header.Add("X-Timestamp", fmt.Sprint(time.Now().UnixMilli()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError.Println("Failed to find EOF in response")
	}

	if resp.StatusCode != http.StatusOK {
		return "", GenerateError(resp, string(respBody))
	}

	sb := strings.Trim(string(respBody), "\\\"")
	return sb, nil
}

func (g *Getter) GetRequest(c *gin.Context) ([]byte, error) {
	var serviceUrl string

	switch g.ServiceName {
	case datastruct.LABORATORY:
		serviceUrl = config.LabServiceURL
	case datastruct.RADIOLOGY:
		serviceUrl = config.RadiologyServiceURL
	case datastruct.PHARMACY:
		serviceUrl = config.PharmacyServiceURL
	default:
		return nil, fmt.Errorf("service %s undefined", g.ServiceName)
	}

	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/request/%s/%s/%s", serviceUrl, g.ServiceName, g.NoIHS, g.RefID),
		nil,
	)
	req.Header.Add("Authorization", c.GetHeader("Authorization"))
	req.Header.Add("X-Timestamp", fmt.Sprint(time.Now().UnixMilli()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogPanic.Panicln("Failed to find EOF in response")
	}

	return respBody, nil
}

func GenerateError(response *http.Response, respBody string) error {
	return fmt.Errorf("error creating request | %s - %s", response.Status, respBody)
}
