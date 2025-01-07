//go:build e2e

package shared

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	argov1alpha1 "github.com/cnoe-io/argocd-api/api/argo/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/entity"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/tests/e2e/container"
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

func TestCoreEndpoints(ctx context.Context, t *testing.T, containerEngine container.Engine, argoBaseUrl, giteaBaseUrl string) {
	TestArgoCDEndpoints(ctx, t, containerEngine, argoBaseUrl)
	TestGiteaEndpoints(ctx, t, containerEngine, giteaBaseUrl)
}

func cleanUp(t *testing.T, containerEngine container.Engine) {
	t.Log("cleaning up environment")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b, err := containerEngine.RunCommand(ctx, fmt.Sprintf("%s ps -aqf name=localdev-control-plane", containerEngine.GetClient()), 10*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while listing docker containers: %s, %s", err, b))

	conts := strings.TrimSpace(string(b))
	if len(conts) == 0 {
		return
	}
	b, err = containerEngine.RunCommand(ctx, fmt.Sprintf("%s container rm -f %s", containerEngine.GetClient(), conts), 60*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while removing docker containers: %s, %s", err, b))

	b, err = containerEngine.RunCommand(ctx, fmt.Sprintf("%s system prune -f", containerEngine.GetClient()), 60*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while pruning system: %s, %s", err, b))

	b, err = containerEngine.RunCommand(ctx, fmt.Sprintf("%s volume prune -f", containerEngine.GetClient()), 60*time.Second)
	assert.NoError(t, err, fmt.Sprintf("error while pruning volumes: %s, %s", err, b))
	t.Log("finished cleaning up container engine environment")
}

// test idpbuilder create
func TestCreateCluster(t *testing.T, containerEngine container.Engine) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer cleanUp(t, containerEngine)

	t.Log("running idpbuilder create")
	b, err := containerEngine.RunIdpCommand(ctx, fmt.Sprintf("%s create --recreate", IdpbuilderBinaryLocation), 0)
	assert.NoError(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))

	TestArgoCDApps(ctx, t, kubeClient, CorePackages)

	argoBaseUrl := fmt.Sprintf("https://argocd.%s:%s", DefaultBaseDomain, DefaultPort)
	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", DefaultBaseDomain, DefaultPort)
	TestCoreEndpoints(ctx, t, containerEngine, argoBaseUrl, giteaBaseUrl)

	TestGiteaRegistry(ctx, t, containerEngine, fmt.Sprintf("gitea.%s", DefaultBaseDomain), DefaultPort)
}

// test idpbuilder create --use-path-routing
func TestCreatePath(t *testing.T, containerEngine container.Engine) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer cleanUp(t, containerEngine)

	t.Log("running idpbuilder create --use-path-routing")
	b, err := containerEngine.RunIdpCommand(ctx, fmt.Sprintf("%s create --use-path-routing", IdpbuilderBinaryLocation), 0)
	assert.NoError(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))

	TestArgoCDApps(ctx, t, kubeClient, CorePackages)

	argoBaseUrl := fmt.Sprintf("https://%s:%s/argocd", DefaultBaseDomain, DefaultPort)
	giteaBaseUrl := fmt.Sprintf("https://%s:%s/gitea", DefaultBaseDomain, DefaultPort)
	TestCoreEndpoints(ctx, t, containerEngine, argoBaseUrl, giteaBaseUrl)

	TestGiteaRegistry(ctx, t, containerEngine, DefaultBaseDomain, DefaultPort)
}

func TestCreatePort(t *testing.T, containerEngine container.Engine) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer cleanUp(t, containerEngine)

	port := "2443"
	t.Logf("running idpbuilder create --port %s", port)
	b, err := containerEngine.RunIdpCommand(ctx, fmt.Sprintf("%s create --port %s", IdpbuilderBinaryLocation, port), 0)
	assert.NoError(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := GetKubeClient()
	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))

	TestArgoCDApps(ctx, t, kubeClient, CorePackages)

	argoBaseUrl := fmt.Sprintf("https://argocd.%s:%s", DefaultBaseDomain, port)
	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", DefaultBaseDomain, port)
	TestCoreEndpoints(ctx, t, containerEngine, argoBaseUrl, giteaBaseUrl)
}

func TestCustomPkg(t *testing.T, containerEngine container.Engine) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	defer cleanUp(t, containerEngine)

	cmdString := "create --package ../../../pkg/controllers/custompackage/test/resources/customPackages/testDir"

	t.Log(fmt.Sprintf("running %s", cmdString))
	b, err := containerEngine.RunIdpCommand(ctx, fmt.Sprintf("%s %s", IdpbuilderBinaryLocation, cmdString), 0)
	assert.NoError(t, err, fmt.Sprintf("error while running create: %s, %s", err, b))

	kubeClient, err := GetKubeClient()

	assert.NoError(t, err, fmt.Sprintf("error while getting client: %s", err))
	if err != nil {
		assert.FailNow(t, "failed creating cluster")
	}

	TestArgoCDApps(ctx, t, kubeClient, CorePackages)

	giteaBaseUrl := fmt.Sprintf("https://gitea.%s:%s", DefaultBaseDomain, DefaultPort)

	expectedApps := map[string]string{
		"my-app":  "argocd",
		"my-app2": "argocd",
	}
	TestArgoCDApps(ctx, t, kubeClient, expectedApps)
	repos, err := GetGiteaRepos(ctx, containerEngine, giteaBaseUrl)
	assert.NoError(t, err)
	expectedRepoNames := map[string]struct{}{
		"idpbuilder-localdev-my-app-app1":  {},
		"idpbuilder-localdev-my-app2-app2": {},
	}

	for i := range repos {
		repo := repos[i]
		_, ok := expectedRepoNames[repo.Name]
		if ok {
			delete(expectedRepoNames, repo.Name)
		}
	}
	assert.Empty(t, expectedRepoNames)
}

// login, build a test image, push, then pull.
func TestGiteaRegistry(ctx context.Context, t *testing.T, containerEngine container.Engine, giteaHost, giteaPort string) {
	t.Log("testing gitea container registry")
	b, err := containerEngine.RunIdpCommand(ctx, fmt.Sprintf("%s get secrets -o json -p gitea", IdpbuilderBinaryLocation), 10*time.Second)
	assert.NoError(t, err)

	secs := make([]entity.Secret, 1)
	err = json.Unmarshal(b, &secs)
	assert.NoError(t, err)

	sec := secs[0]
	user := sec.Username
	pass := sec.Password

	login, err := containerEngine.RunCommand(ctx, fmt.Sprintf("%s login %s:%s -u %s -p %s", containerEngine.GetClient(), giteaHost, giteaPort, user, pass), 10*time.Second)
	require.NoErrorf(t, err, "%s login err: %s", containerEngine.GetClient(), login)

	tag := fmt.Sprintf("%s:%s/giteaadmin/test:latest", giteaHost, giteaPort)

	build, err := containerEngine.RunCommand(ctx, fmt.Sprintf("%s build -f test-dockerfile -t %s .", containerEngine.GetClient(), tag), 10*time.Second)
	require.NoErrorf(t, err, "%s build err: %s", containerEngine.GetClient(), build)

	push, err := containerEngine.RunCommand(ctx, fmt.Sprintf("%s push %s", containerEngine.GetClient(), tag), 10*time.Second)
	require.NoErrorf(t, err, "%s push err: %s", containerEngine.GetClient(), push)

	pull, err := containerEngine.RunCommand(ctx, fmt.Sprintf("%s pull %s", containerEngine.GetClient(), tag), 10*time.Second)
	require.NoErrorf(t, err, "%s pull err: %s", containerEngine.GetClient(), pull)
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

func TestArgoCDEndpoints(ctx context.Context, t *testing.T, containerEngine container.Engine, baseUrl string) {
	t.Log("testing argocd endpoints")
	sessionURL := fmt.Sprintf("%s%s", baseUrl, ArgoCDSessionEndpoint)
	appURL := fmt.Sprintf("%s%s", baseUrl, ArgoCDAppsEndpoint)

	token, err := GetArgoCDSessionToken(ctx, containerEngine, sessionURL)
	assert.Nil(t, err, fmt.Sprintf("getting argocd token: %v", err))

	httpClient := GetHttpClient()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, appURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	var appResp ArgoCDAppResp
	err = SendAndParse(ctx, &appResp, httpClient, req)
	assert.Nil(t, err, fmt.Sprintf("getting argocd applications: %s", err))

	assert.Equal(t, 3, len(appResp.Items), fmt.Sprintf("number of apps do not match: %v", appResp.Items))
}

func TestGiteaEndpoints(ctx context.Context, t *testing.T, containerEngine container.Engine, baseUrl string) {
	t.Log("testing gitea endpoints")
	repos, err := GetGiteaRepos(ctx, containerEngine, baseUrl)
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

func GetGiteaRepos(ctx context.Context, containerEngine container.Engine, baseUrl string) ([]gitea.Repository, error) {
	auth, err := GetBasicAuth(ctx, containerEngine, "gitea-credential")
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

func GetBasicAuth(ctx context.Context, containerEngine container.Engine, name string) (BasicAuth, error) {
	var lastErr error

	for attempt := 0; attempt < 5; attempt++ {
		select {
		case <-ctx.Done():
			return BasicAuth{}, ctx.Err()
		default:
			b, err := containerEngine.RunIdpCommand(ctx, fmt.Sprintf("%s get secrets -o json", IdpbuilderBinaryLocation), 10*time.Second)
			if err != nil {
				lastErr = err
				time.Sleep(httpRetryDelay)
				continue
			}

			out := BasicAuth{}
			secs := make([]entity.Secret, 2)
			if err = json.Unmarshal(b, &secs); err != nil {
				lastErr = err
				time.Sleep(httpRetryDelay)
				continue
			}

			for _, sec := range secs {
				if sec.Name == name {
					out.Password = sec.Password
					out.Username = sec.Username
					break
				}
			}

			if out.Password == "" || out.Username == "" {
				time.Sleep(httpRetryDelay)
				continue
			}

			return out, nil
		}
	}

	return BasicAuth{}, fmt.Errorf("failed after 5 attempts: %w", lastErr)
}

func GetArgoCDSessionToken(ctx context.Context, containerEngine container.Engine, endpoint string) (string, error) {
	auth, err := GetBasicAuth(ctx, containerEngine, "argocd-initial-admin-secret")
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

func isArgoAppSyncedAndHealthy(ctx context.Context, kubeClient client.Client, name, namespace string) (bool, error) {
	app := argov1alpha1.Application{}

	err := kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &app)
	if err != nil {
		return false, err
	}

	return app.Status.Health.Status == "Healthy" && app.Status.Sync.Status == "Synced", nil
}

func GetKubeClient() (client.Client, error) {
	conf, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		return nil, err
	}
	return client.New(conf, client.Options{Scheme: k8s.GetScheme()})
}
