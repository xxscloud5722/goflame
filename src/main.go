package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"goflame/src/host"
	"goflame/src/message"
	"goflame/src/probe"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"time"
)

type Config struct {
	SSL []struct {
		Name        string `yaml:"name"`
		Host        string `yaml:"host"`
		Description string `yaml:"description"`
		Principal   string `yaml:"principal"`
		Message     string `yaml:"message"`
	} `yaml:"ssl"`
	Probe []struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Principal   string `yaml:"principal"`
		Http        []struct {
			Url    string            `yaml:"url"`
			Method string            `yaml:"method"`
			Header map[string]string `yaml:"header"`
			Body   string            `yaml:"body"`
			Form   map[string]string `yaml:"form"`
		} `yaml:"http"`
	} `yaml:"probe"`
	Notice []struct {
		Model string            `yaml:"model"`
		Value string            `yaml:"value"`
		Args  map[string]string `yaml:"args"`
	} `yaml:"notice"`
}

type HostCheckResult struct {
	Host    string
	Expired time.Time
	Message string
}

type ProbeCheckResult struct {
	Name    string
	Url     string
	Message string
}

func (config *Config) sslTask() {
	var result []HostCheckResult
	for _, it := range config.SSL {
		info, err := host.GetHostSSLInfo("https://" + it.Host)
		if err != nil {
			result = append(result, HostCheckResult{
				Host:    it.Host,
				Message: fmt.Sprintf("%s", err),
			})
		} else {
			result = append(result, HostCheckResult{
				Host:    it.Host,
				Expired: info.NotAfter,
			})
		}
	}
	if len(result) == 0 {
		return
	}
	config.sendHostCheckResultNotice(result)
}

func (config *Config) probeTask() {
	var result []ProbeCheckResult
	for _, it := range config.Probe {
		for _, http := range it.Http {
			err := probe.Http(http.Url, http.Method, http.Header, http.Body, http.Form)
			if err != nil {
				result = append(result, ProbeCheckResult{
					Name:    it.Name,
					Url:     http.Url,
					Message: fmt.Sprintf("%s", err),
				})
			}
		}
	}
	if len(result) == 0 {
		return
	}
	config.sendProbeCheckResultNotice(result)
}

func (config *Config) sendHostCheckResultNotice(result []HostCheckResult) {
	for _, it := range config.Notice {
		var text = ""
		var errorCount = 0
		var nearCount = 0
		for _, it := range result {
			if it.Message == "" {
				remainingDay := int32(it.Expired.Sub(time.Now()).Hours() / 24)
				expiredString := it.Expired.Format("2006年01月02日")
				if remainingDay < 7 {
					nearCount++
				}
				text += fmt.Sprintf("> %v ： <font color=\"info\">剩余%v天过期</font><font color=\"comment\">（%v）</font>\n", it.Host, remainingDay, expiredString)
			} else {
				errorCount++
				text += "> " + it.Host + " ： <font color=\"red\">已过期或证书配置错误</font>"
			}
		}
		text = fmt.Sprintf("总共<font color=\"info\">%v张</font>证书，已过期<font color=\"red\">%v张</font>，即将过期<font color=\"warning\">%v张</font>，请相关同事注意。\n%s", len(result), errorCount, nearCount, text)
		if it.Model == "WECHAT" {
			err := message.SendWeChatMessage(it.Value, text)
			if err != nil {
				log.Printf("[推送企业微信失败] %v\n", err)
			}
		}
	}
}
func (config *Config) sendProbeCheckResultNotice(result []ProbeCheckResult) {
	for _, it := range config.Notice {
		var text = ""
		for _, it := range result {
			text += fmt.Sprintf("> **【%s】** 探测（%s）到异常，<font color=\"warning\">异常原因：%s</font>\n", it.Name, it.Url, it.Message)
		}

		if it.Model == "WECHAT" {
			err := message.SendWeChatMessage(it.Value, text)
			if err != nil {
				log.Printf("[推送企业微信失败] %v\n", err)
			}
		}
	}
}

func getConfig(path string) (*Config, error) {
	configYaml, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := Config{}
	err = yaml.Unmarshal(configYaml, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func main() {
	fmt.Println("Load configuration ./application.yaml")
	config, err := getConfig("./application.yaml")
	if err != nil {
		log.Fatalf("Error:%v", err)
	}
	cronTask := cron.New(cron.WithSeconds())
	fmt.Println("Load Cron Task ...")
	// 每天早上执行 9.30 执行
	_, err = cronTask.AddFunc("0 30 9 * * ?", config.sslTask)
	if err != nil {
		log.Fatalf("Error:%v", err)
	}
	// 每天间隔2分钟执行一次
	_, err = cronTask.AddFunc("0 0/2 * * * ?", config.probeTask)
	if err != nil {
		log.Fatalf("Error:%v", err)
	}
	fmt.Println("Application starting Success !")
	cronTask.Start()
	// EPOLL Select
	select {}
}
