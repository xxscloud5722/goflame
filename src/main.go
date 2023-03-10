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
				expiredString := it.Expired.Format("2006???01???02???")
				if remainingDay < 7 {
					nearCount++
				}
				text += fmt.Sprintf("> %v ??? <font color=\"info\">??????%v?????????</font><font color=\"comment\">???%v???</font>\n", it.Host, remainingDay, expiredString)
			} else {
				errorCount++
				text += "> " + it.Host + " ??? <font color=\"red\">??????????????????????????????</font>"
			}
		}
		text = fmt.Sprintf("??????<font color=\"info\">%v???</font>??????????????????<font color=\"red\">%v???</font>???????????????<font color=\"warning\">%v???</font>???????????????????????????\n%s", len(result), errorCount, nearCount, text)
		if it.Model == "WECHAT" {
			err := message.SendWeChatMessage(it.Value, text)
			if err != nil {
				log.Printf("[????????????????????????] %v\n", err)
			}
		}
	}
}
func (config *Config) sendProbeCheckResultNotice(result []ProbeCheckResult) {
	for _, it := range config.Notice {
		var text = ""
		for _, it := range result {
			text += fmt.Sprintf("> **???%s???** ?????????%s???????????????<font color=\"warning\">???????????????%s</font>\n", it.Name, it.Url, it.Message)
		}

		if it.Model == "WECHAT" {
			err := message.SendWeChatMessage(it.Value, text)
			if err != nil {
				log.Printf("[????????????????????????] %v\n", err)
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
	// ?????????????????? 9.30 ??????
	_, err = cronTask.AddFunc("0 30 9 * * ?", config.sslTask)
	if err != nil {
		log.Fatalf("Error:%v", err)
	}
	// ????????????2??????????????????
	_, err = cronTask.AddFunc("0 0/2 * * * ?", config.probeTask)
	if err != nil {
		log.Fatalf("Error:%v", err)
	}
	fmt.Println("Application starting Success !")
	cronTask.Start()
	// EPOLL Select
	select {}
}
