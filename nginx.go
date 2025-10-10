package main

import (
	"fmt"
	"github.com/lincaiyong/log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func writeServerJson(name string, port int) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("location /%s/ {\n\t", name))
	if !config.NoAuthServerMap[name] {
		authCode := strings.ReplaceAll(`set $token_value $cookie_token;
	if ($token_value = "") {
		set $token_value $http_token;
	}
	if ($token_value = "") {
		set $token_value $arg_token;
	}
	if ($token_value != "<token>") {
		return 403;
	}`, "<token>", config.SecretToken)
		sb.WriteString(authCode)
	}
	sb.WriteString(fmt.Sprintf("proxy_pass http://127.0.0.1:%d/;", port))
	sb.WriteString(`	proxy_connect_timeout 5s;
	proxy_read_timeout 60s;
	proxy_send_timeout 60s;
	proxy_next_upstream_tries 3;
	proxy_set_header Host $host;
	proxy_set_header X-Real-IP $remote_addr;
}`)
	_ = os.Mkdir(config.NginxConfDDir, os.ModePerm)
	filePath := path.Join(config.NginxConfDDir, name+".conf")
	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

func getNginxApps() (map[string]int, error) {
	re := regexp.MustCompile(`.+http://127.0.0.1:(\d+)/.+`)
	items, err := os.ReadDir(config.NginxConfDDir)
	if err != nil {
		return nil, fmt.Errorf("fail to read dir: %v", err)
	}
	result := make(map[string]int)
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		name := item.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if strings.HasSuffix(name, ".conf") {
			var b []byte
			if b, err = os.ReadFile(path.Join(config.NginxConfDDir, name)); err != nil {
				return nil, fmt.Errorf("fail to read file: %v", err)
			}
			ret := re.FindStringSubmatch(string(b))
			if len(ret) != 2 {
				log.WarnLog("fail to extract port, skip: %s", name)
				continue
			}
			port, _ := strconv.Atoi(ret[1])
			app := name[:len(name)-len(".conf")]
			result[app] = port
		}
	}
	log.InfoLog("nginx apps: %v", result)
	return result, nil
}

func doReloadNginx(apps map[string]int) error {
	content := `user www-data;
worker_processes auto;
error_log /var/log/nginx/error.log;
events {
	worker_connections 768;
}
http {
	sendfile on;
	tcp_nopush on;
	types_hash_max_size 2048;
	include <nginxdir>/mime.types;
	default_type application/octet-stream;
	ssl_protocols TLSv1 TLSv1.1 TLSv1.2 TLSv1.3; # Dropping SSLv3, ref: POODLE
	ssl_prefer_server_ciphers on;
	access_log /var/log/nginx/access.log;
	<http_server_block>
	<https_server_block>
}`
	httpServerBlock := `server {
		listen 80;
		server_name <domain>;
		client_max_body_size 500M;

		<include_http_conf>
	}`
	httpsServerBlock := `server {
		listen 443 ssl;
		server_name <domain>;
		client_max_body_size 500M;
	
		ssl_certificate <ssldir>/<domain>_chain.pem;
		ssl_certificate_key <ssldir>/<domain>_key.key;

		<include_https_conf>
	}`
	includeHttpConfLines := make([]string, 0)
	includeHttpsConfLines := make([]string, 0)
	for app := range apps {
		if config.HttpServerMap[app] || !config.EnableHttps || runtime.GOOS == "darwin" {
			includeHttpConfLines = append(includeHttpConfLines, fmt.Sprintf("include %s/%s.conf;", config.NginxConfDDir, app))
		} else {
			includeHttpsConfLines = append(includeHttpsConfLines, fmt.Sprintf("include %s/%s.conf;", config.NginxConfDDir, app))
		}
	}
	if len(includeHttpConfLines) > 0 {
		httpServerBlock = strings.ReplaceAll(httpServerBlock, "<ssldir>", config.SSLDir)
		httpServerBlock = strings.ReplaceAll(httpServerBlock, "<include_http_conf>", strings.Join(includeHttpConfLines, "\n\t\t"))
		content = strings.ReplaceAll(content, "<http_server_block>", httpServerBlock)
	} else {
		content = strings.ReplaceAll(content, "<http_server_block>", "")
	}
	if len(includeHttpsConfLines) > 0 {
		httpsServerBlock = strings.ReplaceAll(httpsServerBlock, "<include_https_conf>", strings.Join(includeHttpsConfLines, "\n\t\t"))
		content = strings.ReplaceAll(content, "<https_server_block>", httpsServerBlock)
	} else {
		content = strings.ReplaceAll(content, "<https_server_block>", "")
	}
	content = strings.ReplaceAll(content, "<domain>", config.Domain)
	content = strings.ReplaceAll(content, "<nginxdir>", path.Dir(config.NginxConfFile))
	if runtime.GOOS == "darwin" {
		content = strings.ReplaceAll(content, "user www-data;\n", "")
	}
	err := os.WriteFile(config.NginxConfFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("fail to write nginx.conf: %v", err)
	}
	_ = os.RemoveAll(config.NginxConfDDir)
	_ = os.Mkdir(config.NginxConfDDir, os.ModePerm)
	for app, port := range apps {
		err = writeServerJson(app, port)
		if err != nil {
			return fmt.Errorf("fail to write file: %v", err)
		}
	}
	var out []byte
	log.InfoLog("nginx -s reload")
	out, err = exec.Command("nginx", "-s", "reload").CombinedOutput()
	if err != nil {
		return fmt.Errorf("fail to exec nginx -s reload: %v, %s", err, string(out))
	}
	if runtime.GOOS != "darwin" {
		var output []byte
		output, err = exec.Command("systemctl", "status", "nginx").CombinedOutput()
		if err != nil {
			return fmt.Errorf("fail to systemctl status nginx: %v, %s", err, string(output))
		}
		log.InfoLog("systemctl output: %s", string(output))
	}
	return nil
}
