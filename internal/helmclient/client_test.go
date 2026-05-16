package helmclient

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
	"testing"
)

func readClientSource(t *testing.T) string {
	t.Helper()

	b, err := os.ReadFile("client.go")
	if err != nil {
		t.Fatalf("failed to read client.go: %v", err)
	}
	return string(b)
}

func parseClientSource(t *testing.T, src string) {
	t.Helper()

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "client.go", src, parser.AllErrors); err != nil {
		t.Fatalf("client.go is not valid Go source: %v", err)
	}
}

func TestClientSourceChecks(t *testing.T) {
	src := readClientSource(t)

	tests := []struct {
		name      string
		checkFunc func(string) error
	}{
		{
			name: "client.go parses as valid Go source",
			checkFunc: func(source string) error {
				fset := token.NewFileSet()
				if _, err := parser.ParseFile(fset, "client.go", source, parser.AllErrors); err != nil {
					return fmt.Errorf("client.go is not valid Go source: %w", err)
				}
				return nil
			},
		},
		{
			name: "package name is helmclient",
			checkFunc: func(source string) error {
				re := regexp.MustCompile(`(?m)^\s*package\s+([A-Za-z_]\w*)\s*$`)
				m := re.FindStringSubmatch(source)
				if len(m) != 2 {
					return fmt.Errorf("package declaration not found")
				}
				if got, want := m[1], "helmclient"; got != want {
					return fmt.Errorf("unexpected package name: got %q, want %q", got, want)
				}
				return nil
			},
		},
		{
			name: "contains helm sdk reference",
			checkFunc: func(source string) error {
				if !strings.Contains(source, "helm.sh/helm") {
					return fmt.Errorf("expected Helm SDK import/reference (substring \"helm.sh/helm\") in client.go")
				}
				return nil
			},
		},
		{
			name: "does not contain debug print calls",
			checkFunc: func(source string) error {
				disallowed := []string{
					`fmt.Print("DEBUG:`,
					`fmt.Printf("DEBUG:`,
					`fmt.Println("DEBUG:`,
					`log.Print("DEBUG:`,
					`log.Printf("DEBUG:`,
					`log.Println("DEBUG:`,
				}

				for _, token := range disallowed {
					if strings.Contains(source, token) {
						return fmt.Errorf("disallowed debug print found in client.go: %s", token)
					}
				}

				return nil
			},
		},
		{
			name: "contains v1 operation keyword install",
			checkFunc: func(source string) error {
				if !strings.Contains(strings.ToLower(source), "install") {
					return fmt.Errorf("expected keyword %q to appear in client.go", "install")
				}
				return nil
			},
		},
		{
			name: "contains v1 operation keyword list",
			checkFunc: func(source string) error {
				if !strings.Contains(strings.ToLower(source), "list") {
					return fmt.Errorf("expected keyword %q to appear in client.go", "list")
				}
				return nil
			},
		},
		{
			name: "contains v1 operation keyword upgrade",
			checkFunc: func(source string) error {
				if !strings.Contains(strings.ToLower(source), "upgrade") {
					return fmt.Errorf("expected keyword %q to appear in client.go", "upgrade")
				}
				return nil
			},
		},
		{
			name: "contains v1 operation keyword uninstall",
			checkFunc: func(source string) error {
				if !strings.Contains(strings.ToLower(source), "uninstall") {
					return fmt.Errorf("expected keyword %q to appear in client.go", "uninstall")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.checkFunc(src); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestHelmClientInterface_Contracts(t *testing.T) {
	src := readClientSource(t)

	re := regexp.MustCompile(`(?s)type HelmClient interface \{(.+?)\}`)
	m := re.FindStringSubmatch(src)
	if len(m) != 2 {
		t.Fatalf("HelmClient interface not found in client.go")
	}
	ifaceBody := m[1]

	tests := []struct {
		name      string
		signature string
	}{
		{name: "HelmClient declares Install", signature: "Install("},
		{name: "HelmClient declares ListCharts", signature: "ListCharts("},
		{name: "HelmClient declares Upgrade", signature: "Upgrade("},
		{name: "HelmClient declares Uninstall", signature: "Uninstall("},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(ifaceBody, tt.signature) {
				t.Fatalf("HelmClient interface missing method with signature %q", tt.signature)
			}
		})
	}
}
