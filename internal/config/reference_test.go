package config

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// yamlKeys lists the dotted path of every setting in config.yaml (a map such as the DNS credentials counts as one setting).
func yamlKeys(t reflect.Type, prefix string, out *[]string) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := strings.Split(f.Tag.Get("yaml"), ",")[0]
		if tag == "" || tag == "-" {
			continue
		}
		path := prefix + tag
		if f.Type.Kind() == reflect.Struct {
			yamlKeys(f.Type, path+".", out)
			continue
		}
		*out = append(*out, path)
	}
}

// The reference (docs/11) must describe every setting there is, so it cannot fall behind the program.
func TestConfigurationReferenceDocumentsEverySetting(t *testing.T) {
	b, err := os.ReadFile("../../docs/11-configuration-reference.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(b)
	var keys []string
	yamlKeys(reflect.TypeOf(Config{}), "", &keys)
	if len(keys) < 40 {
		t.Fatalf("expected the whole config, found only %d keys", len(keys))
	}
	for _, k := range keys {
		// a table row starts "| `key`" (or lists several keys in one cell, like cert_file and key_file)
		if !strings.Contains(doc, "`"+k+"`") {
			t.Errorf("docs/11-configuration-reference.md does not describe %s", k)
		}
	}
	// ... and the flags and environment variables of the program
	for _, w := range []string{"--config", "--data", "--http-addr", "--https-addr", "--log-level", "--log-file", "--no-browser", "OPEN_CAPTION_CONFIG", "OPEN_CAPTION_DATA", "OPEN_CAPTION_HTTP_ADDR", "OPEN_CAPTION_HTTPS_ADDR", "OPEN_CAPTION_LOG_LEVEL", "OPEN_CAPTION_LOG"} {
		if !strings.Contains(doc, w) {
			t.Errorf("the reference does not mention %s", w)
		}
	}
}

// The complete example in the reference must load without a single problem.
func TestConfigurationReferenceExampleLoads(t *testing.T) {
	b, err := os.ReadFile("../../docs/11-configuration-reference.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(b)
	i := strings.LastIndex(doc, "```yaml\n")
	if i < 0 {
		t.Fatal("no example")
	}
	ex := doc[i+len("```yaml\n"):]
	ex = ex[:strings.Index(ex, "```")]
	dir := t.TempDir()
	creds := dir + "/dns.env"
	os.WriteFile(creds, []byte("EASYDNS_TOKEN=x\nEASYDNS_KEY=y\n"), 0o600)
	ex = strings.Replace(ex, "/etc/open-caption/dns.env", creds, 1)
	p := dir + "/config.yaml"
	os.WriteFile(p, []byte(ex), 0o600)
	c, err := Load(p)
	if err != nil {
		t.Fatalf("the example in the reference does not load: %v", err)
	}
	if !c.HTTPS.Enabled || !c.HTTPS.ACME.Enabled || c.HTTPS.ACME.Domain != "captions.mytheater.org" || !c.Access.LocalAutoLogin || c.Defaults.Look.Font != "hyperlegible" {
		t.Errorf("the example was read wrongly: %+v", c.HTTPS)
	}
}

// The settings shown in the step-by-step guide must be valid too (it only shows the https part; passwords are added here).
func TestHTTPSGuideConfigurationLoads(t *testing.T) {
	b, err := os.ReadFile("../../docs/https-certificate-guide.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(b)
	i := strings.Index(doc, "```yaml\n")
	if i < 0 {
		t.Fatal("no example")
	}
	ex := doc[i+len("```yaml\n"):]
	ex = ex[:strings.Index(ex, "```")]
	dir := t.TempDir()
	creds := dir + "/dns.env"
	os.WriteFile(creds, []byte("EASYDNS_TOKEN=x\nEASYDNS_KEY=y\n"), 0o600)
	ex = "admin_password: a\noperator_password: b\n" + strings.Replace(ex, "/etc/open-caption/dns.env", creds, 1)
	p := dir + "/config.yaml"
	os.WriteFile(p, []byte(ex), 0o600)
	c, err := Load(p)
	if err != nil {
		t.Fatalf("the example in the guide does not load: %v", err)
	}
	if !c.HTTPS.ACME.Enabled || c.HTTPS.ACME.Directory != ACMEDirectoryStaging || c.HTTPS.ACME.DNS.Provider != "easydns" {
		t.Errorf("%+v", c.HTTPS.ACME)
	}
}
