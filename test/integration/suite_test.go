package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// PushIntegrationSuite provides a test suite for push command integration tests
type PushIntegrationSuite struct {
	suite.Suite
	// Mock registry configuration
	registryHost   string
	registryPort   string
	registryURL    string
	testImageName  string
	testImageTag   string
	cleanup        []func() error
}

// SetupSuite initializes the test environment before running the test suite
func (suite *PushIntegrationSuite) SetupSuite() {
	// Initialize test registry configuration
	suite.registryHost = "localhost"
	suite.registryPort = "5000"
	suite.registryURL = suite.registryHost + ":" + suite.registryPort
	suite.testImageName = "test-image"
	suite.testImageTag = "integration-test"

	// Initialize cleanup functions slice
	suite.cleanup = make([]func() error, 0)

	suite.T().Log("Integration test suite setup completed")
	suite.T().Logf("Test registry: %s", suite.registryURL)
	suite.T().Logf("Test image: %s/%s:%s", suite.registryURL, suite.testImageName, suite.testImageTag)
}

// TearDownSuite cleans up test resources after the test suite completes
func (suite *PushIntegrationSuite) TearDownSuite() {
	// Run all cleanup functions in reverse order
	for i := len(suite.cleanup) - 1; i >= 0; i-- {
		if err := suite.cleanup[i](); err != nil {
			suite.T().Errorf("Cleanup function failed: %v", err)
		}
	}

	suite.T().Log("Integration test suite cleanup completed")
}

// SetupTest runs before each individual test
func (suite *PushIntegrationSuite) SetupTest() {
	suite.T().Logf("Starting test: %s", suite.T().Name())
}

// TearDownTest runs after each individual test
func (suite *PushIntegrationSuite) TearDownTest() {
	suite.T().Logf("Completed test: %s", suite.T().Name())
}

// addCleanup adds a cleanup function to be executed during teardown
func (suite *PushIntegrationSuite) addCleanup(cleanup func() error) {
	suite.cleanup = append(suite.cleanup, cleanup)
}

// getTestImageURL returns the full URL for the test image
func (suite *PushIntegrationSuite) getTestImageURL() string {
	return suite.registryURL + "/" + suite.testImageName + ":" + suite.testImageTag
}

// getTestImageURLWithCustomTag returns the test image URL with a custom tag
func (suite *PushIntegrationSuite) getTestImageURLWithCustomTag(tag string) string {
	return suite.registryURL + "/" + suite.testImageName + ":" + tag
}

// TestPushIntegrationSuite runs the integration test suite
func TestPushIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PushIntegrationSuite))
}