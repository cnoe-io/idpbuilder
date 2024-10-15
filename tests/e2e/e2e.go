//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	argov1alpha1 "github.com/cnoe-io/argocd-api/api/argo/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/get"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	IdpbuilderBinaryLocation = "../../../idpbuilder"
	DefaultPort              = "8443"
	DefaultBaseDomain        = "cnoe.localtest.me"
	ArgoCDSessionEndpoint    = "/api/v1/session"
	ArgoCDAppsEndpoint       = "/api/v1/applications"
	GiteaSessionEndpoint     = "/api/v1/users/%s/tokens"
	GiteaUserEndpoint        = "/api/v1/users/%s"
	GiteaRepoEndpoint        = "/api/v1/repos/search"

	httpRetryDelay   = 5 * time.Second
	httpRetryTimeout = 300 * time.Second
)

var (
	// CorePackages is a map of argocd app name to its namespace.
	CorePackages = map[string]string{
		"argocd": "argocd",
		"nginx":  "argocd",
		"gitea":  "argocd",
	}
)

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ArgoCDAuthResponse struct {
	Token string `json:"token"`
}

type ArgoCDAppResp struct {
	Items []argov1alpha1.Application
}

type GiteaSearchRepoResponse struct {
	Ok   bool
	Data []gitea.Repository
}

func GetHttpClient() *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &http.Client{Transport: tr}
}

func TestCoreEndpoints(ctx context.Context, t *testing.T, argoBaseUrl, giteaBaseUrl string) {
	TestArgoCDEndpoints(ctx, t, argoBaseUrl)
	TestGiteaEndpoints(ctx, t, giteaBaseUrl)
}

func RunCommand(ctx context.Context, command string, timeout time.Duration) ([]byte, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmds := strings.Split(command, " ")
	if len(cmds) == 0 {
		return nil, fmt.Errorf("supply at least one command")
	}
	binary := cmds[0]
	args := make([]string, 0, len(cmds)-1)
	if len(cmds) > 1 {
		args = append(args, cmds[1:]...)
	}

	c := exec.CommandContext(cmdCtx, binary, args...)
	b, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error while running %s: %s, %s", command, err, b)
	}

	return b, nil
}

func SendAndParse(ctx context.Context, target any, httpClient *http.Client, req *http.Request) error {
	sendCtx, cancel := context.WithTimeout(ctx, httpRetryTimeout)
	defer cancel()
	var bodyBytes []byte
	if req.Body != nil {
		b, bErr := io.ReadAll(req.Body)
		if bErr != nil {
			return fmt.Errorf("failed copying http request body: %w", bErr)
		}
		bodyBytes = b
	}

	for {
		select {
		case <-sendCtx.Done():
			return fmt.Errorf("timedout")
		default:
			if req.Body != nil {
				b := append(make([]byte, 0, len(bodyBytes)), bodyBytes...)
				req.Body = io.NopCloser(bytes.NewBuffer(b))
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				fmt.Println("failed running http request: ", err)
				time.Sleep(httpRetryDelay)
				continue
			}

			defer resp.Body.Close()

			respB, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("failed reading http response body: ", err)
				time.Sleep(httpRetryDelay)
				continue
			}

			err = json.Unmarshal(respB, target)
			if err != nil {
				fmt.Println("failed parsing response body: ", err, "\n", string(respB))
				time.Sleep(httpRetryDelay)
				continue
			}
			return nil
		}
	}
}

func TestGiteaEndpoints(ctx context.Context, t *testing.T, baseUrl string) {
	t.Log("testing gitea endpoints")
	repos, err := GetGiteaRepos(ctx, baseUrl)
	assert.Nil(t, err)

	assert.Equal(t, 3, len(repos))
	expectedRepoNames := map[string]struct{}{
		"idpbuilder-localdev-gitea":  {},
		"idpbuilder-localdev-nginx":  {},
		"idpbuilder-localdev-argocd": {},
	}

	for i := range repos {
		_, ok := expectedRepoNames[repos[i].Name]
		if ok {
			delete(expectedRepoNames, repos[i].Name)
		}
	}
	assert.Equal(t, 0, len(expectedRepoNames))
}

func GetGiteaRepos(ctx context.Context, baseUrl string) ([]gitea.Repository, error) {
	auth, err := GetBasicAuth(ctx, "gitea-credential")
	if err != nil {
		return nil, fmt.Errorf("getting gitea credentials %w", err)
	}

	token, err := GetGiteaSessionToken(ctx, auth, baseUrl)
	if err != nil {
		return nil, fmt.Errorf("getting gitea token %w", err)
	}

	userEP := fmt.Sprintf("%s%s", baseUrl, fmt.Sprintf(GiteaUserEndpoint, auth.Username))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userEP, nil)
	if err != nil {
		return nil, fmt.Errorf("creating new request %w", err)
	}

	httpClient := GetHttpClient()
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	user := gitea.User{}
	err = SendAndParse(ctx, &user, httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("getting user info %w", err)
	}

	repos := GiteaSearchRepoResponse{}
	repoEp := fmt.Sprintf("%s%s", baseUrl, GiteaRepoEndpoint)
	repoReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, repoEp, nil)
	err = SendAndParse(ctx, &repos, httpClient, repoReq)
	if err != nil {
		return nil, fmt.Errorf("getting gitea repositories %w", err)
	}

	return repos.Data, nil
}

func GetGiteaSessionToken(ctx context.Context, auth BasicAuth, baseUrl string) (string, error) {
	httpClient := GetHttpClient()
	sessionEP := fmt.Sprintf("%s%s", baseUrl, fmt.Sprintf(GiteaSessionEndpoint, auth.Username))

	sb := []byte(fmt.Sprintf(`{"name":"%d"}`, time.Now().Unix()))
	sessionReq, err := http.NewRequestWithContext(ctx, http.MethodPost, sessionEP, bytes.NewBuffer(sb))
	if err != nil {
		return "", fmt.Errorf("reating new request for session: %w", err)
	}

	sessionReq.SetBasicAuth(auth.Username, auth.Password)
	sessionReq.Header.Set("Content-Type", "application/json")

	var sess gitea.AccessToken
	err = SendAndParse(ctx, &sess, httpClient, sessionReq)
	if err != nil {
		return "", err
	}

	if sess.Token == "" {
		return "", fmt.Errorf("received empty token")
	}
	return sess.Token, nil
}

func TestArgoCDEndpoints(ctx context.Context, t *testing.T, baseUrl string) {
	t.Log("testing argocd endpoints")
	sessionURL := fmt.Sprintf("%s%s", baseUrl, ArgoCDSessionEndpoint)
	appURL := fmt.Sprintf("%s%s", baseUrl, ArgoCDAppsEndpoint)

	token, err := GetArgoCDSessionToken(ctx, sessionURL)
	assert.Nil(t, err, fmt.Sprintf("getting argocd token: %v", err))

	httpClient := GetHttpClient()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, appURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	var appResp ArgoCDAppResp
	err = SendAndParse(ctx, &appResp, httpClient, req)
	assert.Nil(t, err, fmt.Sprintf("getting argocd applications: %s", err))

	assert.Equal(t, 3, len(appResp.Items), fmt.Sprintf("number of apps do not match: %v", appResp.Items))
}

func GetBasicAuth(ctx context.Context, name string) (BasicAuth, error) {

	b, err := RunCommand(ctx, fmt.Sprintf("%s get secrets -o json", IdpbuilderBinaryLocation), 10*time.Second)
	if err != nil {
		return BasicAuth{}, err
	}
	out := BasicAuth{}

	secs := make([]get.TemplateData, 2)
	err = json.Unmarshal(b, &secs)
	if err != nil {
		return BasicAuth{}, err
	}

	for i := range secs {
		if secs[i].Name == name {
			out.Password = secs[i].Data["password"]
			out.Username = secs[i].Data["username"]
			break
		}
	}
	if out.Password == "" || out.Username == "" {
		return BasicAuth{}, fmt.Errorf("could not find argocd or gitea credentials: %s", b)
	}
	return out, nil
}

func GetArgoCDSessionToken(ctx context.Context, endpoint string) (string, error) {
	auth, err := GetBasicAuth(ctx, "argocd-initial-admin-secret")
	if err != nil {
		return "", err
	}
	httpClient := GetHttpClient()

	authJ, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(authJ))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	var tokenResp ArgoCDAuthResponse
	err = SendAndParse(ctx, &tokenResp, httpClient, req)
	if err != nil {
		return "", err
	}

	if tokenResp.Token == "" {
		return "", fmt.Errorf("received token is empty")
	}

	return tokenResp.Token, nil
}

func TestArgoCDApps(ctx context.Context, t *testing.T, kubeClient client.Client, apps map[string]string) {
	done := false
	for !done {
		select {
		case <-ctx.Done():
			return
		default:
			for k := range apps {
				ns := apps[k]
				t.Logf("checking argocd app %s in %s ns", k, ns)
				ready, argoErr := isArgoAppSyncedAndHealthy(ctx, kubeClient, k, ns)
				if argoErr != nil {
					t.Logf("error when checking ArgoCD app health: %s", argoErr)
					continue
				}
				if ready {
					t.Logf("app %s ready", k)
					delete(apps, k)
				}
			}
			if len(apps) != 0 {
				t.Logf("waiting for apps to be ready")
				time.Sleep(httpRetryDelay)
				continue
			}
			done = true
			t.Log("all argocd apps healthy")
		}
	}
}

func isArgoAppSyncedAndHealthy(ctx context.Context, kubeClient client.Client, name, namespace string) (bool, error) {
	app := argov1alpha1.Application{}

	err := kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &app)
	if err != nil {
		return false, err
	}

	if app.Status.Health.Status == "Healthy" && app.Status.Sync.Status == "Synced" {
		return true, nil
	}

	return false, nil
}

func GetKubeClient() (client.Client, error) {
	conf, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		return nil, err
	}
	return client.New(conf, client.Options{Scheme: k8s.GetScheme()})
}

// login, build a test image, push, then pull.
func TestGiteaRegistry(ctx context.Context, t *testing.T, cmd, giteaHost, giteaPort string) {
	t.Log("testing gitea container registry")
	b, err := RunCommand(ctx, fmt.Sprintf("%s get secrets -o json -p gitea", IdpbuilderBinaryLocation), 10*time.Second)
	assert.NoError(t, err)

	secs := make([]get.TemplateData, 2)
	err = json.Unmarshal(b, &secs)
	assert.NoError(t, err)

	sec := secs[0]
	user := sec.Data["username"]
	pass := sec.Data["password"]

	login, err := RunCommand(ctx, fmt.Sprintf("%s login %s:%s -u %s -p %s", cmd, giteaHost, giteaPort, user, pass), 10*time.Second)
	require.NoErrorf(t, err, "%s login err: %s", cmd, login)

	tag := fmt.Sprintf("%s:%s/giteaadmin/test:latest", giteaHost, giteaPort)

	build, err := RunCommand(ctx, fmt.Sprintf("%s build -f test-dockerfile -t %s .", cmd, tag), 10*time.Second)
	require.NoErrorf(t, err, "%s build err: %s", cmd, build)

	push, err := RunCommand(ctx, fmt.Sprintf("%s push %s", cmd, tag), 10*time.Second)
	require.NoErrorf(t, err, "%s push err: %s", cmd, push)

	pull, err := RunCommand(ctx, fmt.Sprintf("%s pull %s", cmd, tag), 10*time.Second)
	require.NoErrorf(t, err, "%s pull err: %s", cmd, pull)
}
