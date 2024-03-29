package e2e

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	keycloakv1alpha1 "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	"github.com/christianwoehrle/keycloakclient-controller/pkg/common"
	"github.com/christianwoehrle/keycloakclient-controller/pkg/model"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	secondClientName = "test-client-second"
	authZClientName  = "test-client-authz"
)

var ErrDeprecatedClientSecretFound = errors.New("deprecated client secret found")

var _ = Describe("KeycloakClient", func() {

	BeforeEach(func() {
		err := prepareExternalKeycloaksCR()
		Expect(err).To(BeNil())
		prepareKeycloakRealmCR()
		tearDownKeycloakClients()
	})
	AfterEach(func() {
		err := tearDownExternalKeycloaksRealmCR()
		Expect(err).To(BeNil())
		err = tearDownExternalKeycloaksCR()
		Expect(err).To(BeNil())
		tearDownKeycloakClients()
	})

	Describe("keycloakClientBasicTest", func() {
		BeforeEach(func() {
			prepareKeycloakClientCR()
		})
		It("basic client can be created", func() {
			err := WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)
			Expect(err).To(BeNil())
		})

	})
	Describe("externalKeycloakClientBasicTest", func() {
		BeforeEach(func() {
			prepareExternalKeycloakClientCR()
		})
		It("external client can be created and moves to ready", func() {
			err := WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)
			Expect(err).To(BeNil())
		})
	})
	Describe("keycloakClientAuthZSettingsTest", func() {
		BeforeEach(func() {
			prepareKeycloakClientAuthZCR()
		})
		It("test basic client", func() {
			err := keycloakClientAuthZTest()
			Expect(err).To(BeNil())
		})
	})
	Describe("keycloakClientRolesTest", func() {
		BeforeEach(func() {
			prepareKeycloakClientAuthZCR()
		})
		It("keycloakClientRolesTest", func() {
			err := keycloakClientRolesTest()
			Expect(err).To(BeNil())
		})
	})
	Describe("keycloakClientDefaultRolesTest", func() {
		It("test basic client", func() {
			err := keycloakClientDefaultRolesTest()
			Expect(err).To(BeNil())
		})
	})

	Describe("keycloakClientScopeMappingsTest", func() {
		BeforeEach(func() {
			prepareKeycloakClientWithRolesCR()
		})
		It("test basic client", func() {
			err := keycloakClientScopeMappingsTest()
			Expect(err).To(BeNil())
		})
	})

	Describe("keycloakClientDeprecatedClientSecretTest", func() {
		It("test basic client", func() {
			err := keycloakClientDeprecatedClientSecretTest()
			Expect(err).To(BeNil())
		})
	})

	Describe("keycloakClientServiceAccountRealmRolesTest", func() {
		It("test basic client", func() {
			err := keycloakClientServiceAccountRealmRolesTest()
			Expect(err).To(BeNil())
		})
	})

})

func getKeycloakClientCR() *keycloakv1alpha1.KeycloakClient {
	k8sName := testKeycloakClientCRName
	id := testKeycloakClientCRName
	labels := CreateLabel(keycloakNamespace)

	return &keycloakv1alpha1.KeycloakClient{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sName,
			Namespace: keycloakNamespace,
			Labels:    labels,
		},
		Spec: keycloakv1alpha1.KeycloakClientSpec{
			RealmSelector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Client: &keycloakv1alpha1.KeycloakAPIClient{
				ID:                        id,
				ClientID:                  id,
				Name:                      id,
				SurrogateAuthRequired:     false,
				Enabled:                   true,
				BaseURL:                   "https://operator-test.url/client-base-url",
				AdminURL:                  "https://operator-test.url/client-admin-url",
				RootURL:                   "https://operator-test.url/client-root-url",
				Description:               "Client used within operator tests",
				WebOrigins:                []string{"https://operator-test.url"},
				BearerOnly:                false,
				ConsentRequired:           false,
				StandardFlowEnabled:       true,
				ImplicitFlowEnabled:       false,
				DirectAccessGrantsEnabled: true,
				ServiceAccountsEnabled:    false,
				PublicClient:              true,
				FrontchannelLogout:        false,
				Protocol:                  "openid-connect",
				FullScopeAllowed:          &[]bool{true}[0],
				NodeReRegistrationTimeout: -1,
				DefaultClientScopes:       []string{"profile"},
				OptionalClientScopes:      []string{"microprofile-jwt"},
			},
		},
	}
}

func getKeycloakClientAuthZCR() *keycloakv1alpha1.KeycloakClient {
	k8sName := testAuthZKeycloakClientCRName
	id := authZClientName
	labels := CreateLabel(keycloakNamespace)

	audioResourceType := "urn:" + id + ":resources:audio"
	imageResourceType := "urn:" + id + ":resources:image"

	return &keycloakv1alpha1.KeycloakClient{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sName,
			Namespace: keycloakNamespace,
			Labels:    labels,
		},
		Spec: keycloakv1alpha1.KeycloakClientSpec{
			RealmSelector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Client: &keycloakv1alpha1.KeycloakAPIClient{
				ID:                           id,
				ClientID:                     id,
				Name:                         id,
				Description:                  "AuthZ Client used within operator tests",
				PublicClient:                 false,
				ServiceAccountsEnabled:       true,
				AuthorizationServicesEnabled: true,
				AuthorizationSettings: &keycloakv1alpha1.KeycloakResourceServer{
					Resources: []keycloakv1alpha1.KeycloakResource{
						{
							Name: "Audio Resource",
							Uris: []string{"/audio"},
							Type: audioResourceType,
							Scopes: []apiextensionsv1.JSON{
								{Raw: []byte(`{"name": "audio:listen"}`)},
							},
						},
						{
							Name: "Image Resource",
							Uris: []string{"/image"},
							Type: imageResourceType,
							Scopes: []apiextensionsv1.JSON{
								{Raw: []byte(`{"name": "image:create"}`)},
								{Raw: []byte(`{"name": "image:read"}`)},
								{Raw: []byte(`{"name": "image:delete"}`)},
							},
						},
					},
					Policies: []keycloakv1alpha1.KeycloakPolicy{
						{
							Name:        "Role Policy",
							Description: "A policy that is role based",
							Type:        "role",
							Logic:       "POSITIVE",
							Config: map[string]string{
								"roles": "[{\"id\":\"" + id + "/uma_protection\",\"required\":true}]",
							},
						},
						{
							Name:             "Aggregate Policy",
							Description:      "A policy that is an aggregate",
							Type:             "aggregate",
							Logic:            "POSITIVE",
							DecisionStrategy: "AFFIRMATIVE",
							Config: map[string]string{
								"applyPolicies": "[\"Role Policy\",\"Deny Policy\"]",
							},
						},
						{
							Name:             "Audio Permission",
							Description:      "An audio permission description",
							Type:             "resource",
							DecisionStrategy: "AFFIRMATIVE",
							Config: map[string]string{
								"defaultResourceType": audioResourceType,
								"default":             "true",
								"applyPolicies":       "[\"Time Policy\"]",
								"scopes":              "[\"audio:listen\"]",
							},
						},
						{
							Name:             "Image Permission",
							Description:      "An image permission description",
							Type:             "scope",
							DecisionStrategy: "UNANIMOUS",
							Config: map[string]string{
								"applyPolicies": "[\"Deny Policy\"]",
								"scopes":        "[\"image:delete\"]",
							},
						},
						{
							Name:        "Deny Policy",
							Description: "A policy that is JS based",
							Type:        "js",
							Config: map[string]string{
								"code": "$evaluation.deny();",
							},
						},
						{
							Name:        "Time Policy",
							Description: "A policy that grants access between 3 and 5 PM",
							Type:        "time",
							Logic:       "POSITIVE",
							Config: map[string]string{
								"hour":    "15",
								"hourEnd": "17",
							},
						},
					},
					Scopes: []keycloakv1alpha1.KeycloakScope{
						{Name: "audio:listen"},
						{Name: "image:create"},
						{Name: "image:read"},
						{Name: "image:delete"},
					},
				},
			},
		},
	}
}

func getKeycloakClientWithServiceAccount() *keycloakv1alpha1.KeycloakClient {
	keycloakClientCR := getKeycloakClientCR()
	keycloakClientCR.Spec.Client.PublicClient = false
	keycloakClientCR.Spec.Client.ServiceAccountsEnabled = true
	keycloakClientCR.Spec.ServiceAccountRealmRoles = []string{"realmRoleA", "realmRoleB"}
	keycloakClientCR.Spec.ServiceAccountClientRoles = map[string][]string{secondClientName: {"a", "b"}}
	return keycloakClientCR
}

func prepareKeycloakClientCR() error {
	keycloakClientCR := getKeycloakClientCR()
	_, err := CreateKeycloakClient(keycloakClientCR)
	return err
}

func prepareExternalKeycloakClientCR() error {
	keycloakClientCR := getKeycloakClientCR()
	_, err := CreateKeycloakClient(keycloakClientCR)
	return err
}

func prepareKeycloakClientAuthZCR() error {
	keycloakClientCR := getKeycloakClientAuthZCR()
	_, err := CreateKeycloakClient(keycloakClientCR)
	return err
}

func prepareKeycloakClientWithServiceAccount() (*keycloakv1alpha1.KeycloakClient, error) {
	keycloakClientCR := getKeycloakClientWithServiceAccount()
	return CreateKeycloakClient(keycloakClientCR)
}

func keycloakClientBasicTest() error {
	return WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)
}

func externalKeycloakClientBasicTest() error {
	return WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)
}

func keycloakClientAuthZTest() error {
	return WaitForClientToBeReady(keycloakNamespace, testAuthZKeycloakClientCRName)
}

func keycloakClientDeprecatedClientSecretTest() error {
	client := getKeycloakClientCR()
	secret := model.DeprecatedClientSecret(client)

	// create client secret using client ID, i.e., keycloak-client-secret-<CLIENT_ID>
	err := CreateSecret(secret)
	if err != nil {
		return err
	}

	// create client
	client, err = CreateKeycloakClient(client)
	if err != nil {
		return err
	}
	err = WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)
	if err != nil {
		return err
	}

	// verify client secret removal in secondary resources
	_, exists := client.Status.SecondaryResources[secret.Name]
	if exists {
		return errors.Wrap(ErrDeprecatedClientSecretFound, secret.Name)
	}

	// verify client secret removal
	var retrievedSecret v1.Secret
	err = GetNamespacedSecret(keycloakNamespace, secret.Name, &retrievedSecret)
	if !apierrors.IsNotFound(err) {
		return err
	}

	return nil
}

func keycloakClientRolesTest() error {
	// create
	client := getKeycloakClientCR()

	client.Spec.Roles = []keycloakv1alpha1.RoleRepresentation{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	client, err := CreateKeycloakClient(client)
	if err != nil {
		return err
	}
	err = WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)
	if err != nil {
		return err
	}

	// update client: delete/rename/leave/add role
	keycloakCR, err := getDeployedKeycloakCR(keycloakNamespace)
	authenticatedClient, err := MakeAuthenticatedClient(*keycloakCR)
	if err != nil {
		return err
	}
	bID, err := getClientRoleID(authenticatedClient, testKeycloakClientCRName, "b")
	if err != nil {
		return err
	}
	client, err = GetNamespacedKeycloakClient(keycloakNamespace, testKeycloakClientCRName)
	if err != nil {
		return err
	}
	client.Spec.Roles = []keycloakv1alpha1.RoleRepresentation{{ID: bID, Name: "b2"}, {Name: "c"}, {Name: "d"}}
	_, err = UpdateKeycloakClient(keycloakNamespace, client)
	if err != nil {
		return err
	}
	// check role presence directly as a "ready" CR might just be stale
	err = waitForClientRoles(*keycloakCR, client, client.Spec.Roles)
	if err != nil {
		return err
	}
	return WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)
}

func keycloakClientDefaultRolesTest() error {
	// create
	client := getKeycloakClientCR()
	client.Spec.Roles = []keycloakv1alpha1.RoleRepresentation{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	client.Spec.Client.DefaultRoles = []string{"a", "b"}
	client, err := CreateKeycloakClient(client)
	if err != nil {
		return err
	}
	err = WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)
	if err != nil {
		return err
	}

	keycloakCR, err := getDeployedKeycloakCR(keycloakNamespace)
	if err != nil {
		return err
	}

	// are roles "a" and "b" the ONLY default roles for this client?
	err = waitForDefaultClientRoles(*keycloakCR, client, "a", "b")
	if err != nil {
		return err
	}

	// update default client roles
	client, err = GetNamespacedKeycloakClient(keycloakNamespace, testKeycloakClientCRName)
	if err != nil {
		return err
	}
	client.Spec.Client.DefaultRoles = []string{"b", "c"}
	client, err = UpdateKeycloakClient(keycloakNamespace, client)
	if err != nil {
		return err
	}

	// are roles "b" and "c" the ONLY default roles for this client?
	err = waitForDefaultClientRoles(*keycloakCR, client, "b", "c")
	if err != nil {
		return err
	}

	return nil
}

func getClientRoleID(authenticatedClient common.KeycloakInterface, clientName, roleName string) (string, error) {
	retrievedRoles, err := authenticatedClient.ListClientRoles(clientName, realmName)
	if err != nil {
		return "", err
	}
	return getRole(retrievedRoles, roleName), nil
}

func getRole(retrievedRoles []keycloakv1alpha1.RoleRepresentation, roleName string) string {
	for _, role := range retrievedRoles {
		if role.Name == roleName {
			return role.ID
		}
	}
	return ""
}

func waitForClientRoles(keycloakCR keycloakv1alpha1.Keycloak, clientCR *keycloakv1alpha1.KeycloakClient, expected []keycloakv1alpha1.RoleRepresentation) error {
	return WaitForConditionWithClient(keycloakCR, func(authenticatedClient common.KeycloakInterface) error {
		roles, err := authenticatedClient.ListClientRoles(clientCR.Spec.Client.ID, realmName)
		if err != nil {
			return err
		}

		fail := false
		if len(roles) != len(expected) {
			fail = true
		} else {
			for _, expectedRole := range expected {
				found := false
				for _, role := range roles {
					if role.Name == expectedRole.Name && (expectedRole.ID == "" || role.ID == expectedRole.ID) {
						found = true
						break
					}
				}
				if !found {
					fail = true
					break
				}
			}
		}

		if fail {
			return errors.Errorf("role names not as expected:\n"+
				"expected: %v\n"+
				"actual  : %v",
				expected, roles)
		}
		return nil
	})
}

func waitForDefaultClientRoles(keycloakCR keycloakv1alpha1.Keycloak, clientCR *keycloakv1alpha1.KeycloakClient, expectedRoleNames ...string) error {
	return WaitForConditionWithClient(keycloakCR, func(authenticatedClient common.KeycloakInterface) error {
		fmt.Println("waitForDefaultClientRoles")
		fail := false

		realm, err := authenticatedClient.GetRealm(realmName)
		if err != nil {
			return err
		}

		defaultRoles, err := authenticatedClient.ListRealmRoleClientRoleComposites(realmName, realm.Spec.Realm.DefaultRole.ID, clientCR.Spec.Client.ID)
		if err != nil {
			return err
		}

		// check if roles and defaultRoles equal
		if len(expectedRoleNames) != len(defaultRoles) {
			fail = true
		}
		for _, expected := range expectedRoleNames {
			found := false
			for _, actual := range defaultRoles {
				if expected == actual.Name {
					found = true
					break
				}
			}
			if !found {
				fail = true
			}
		}

		if fail {
			return errors.Errorf("default roles not as expected:\n"+
				"expected: %v\n"+
				"actual  : %v",
				expectedRoleNames, defaultRoles)
		}

		return nil
	})
}

func prepareKeycloakClientWithRolesCR() (*keycloakv1alpha1.KeycloakClient, error) {
	keycloakClientCR := getKeycloakClientCR().DeepCopy()
	keycloakClientCR.Spec.Roles = []keycloakv1alpha1.RoleRepresentation{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	keycloakClientCR.Name = testSecondKeycloakClientCRName
	keycloakClientCR.Spec.Client.ID = secondClientName
	keycloakClientCR.Spec.Client.ClientID = secondClientName
	keycloakClientCR.Spec.Client.Name = secondClientName
	keycloakClientCR.Spec.Client.WebOrigins = []string{"https://operator-test-second.url"}
	return CreateKeycloakClient(keycloakClientCR)
}

func getKeycloakClientWithScopeMappingsCR(authenticatedClient common.KeycloakInterface,
	realmRoleNames []string, secondClientRoleNames []string) (*keycloakv1alpha1.KeycloakClient, error) {
	client := getKeycloakClientCR()
	mappings, err := getKeycloakClientScopeMappings(authenticatedClient, realmRoleNames, secondClientRoleNames)
	if err != nil {
		return nil, err
	}
	client.Spec.ScopeMappings = mappings
	return client, nil
}

func getKeycloakClientScopeMappings(authenticatedClient common.KeycloakInterface, realmRoleNames []string,
	secondClientRoleNames []string) (*keycloakv1alpha1.MappingsRepresentation, error) {
	var scopeMappings = keycloakv1alpha1.MappingsRepresentation{
		ClientMappings: make(map[string]keycloakv1alpha1.ClientMappingsRepresentation),
	}
	for _, roleName := range realmRoleNames {
		scopeMappings.RealmMappings = append(scopeMappings.RealmMappings, keycloakv1alpha1.RoleRepresentation{
			ID:   roleName,
			Name: roleName,
		})
	}

	secondClient := keycloakv1alpha1.ClientMappingsRepresentation{ID: secondClientName, Client: secondClientName}
	for _, roleName := range secondClientRoleNames {
		roleID, err := getClientRoleID(authenticatedClient, secondClientName, roleName)
		if err != nil {
			return nil, err
		}
		secondClient.Mappings = append(secondClient.Mappings, keycloakv1alpha1.RoleRepresentation{
			ID:   roleID,
			Name: roleName,
		})
	}
	scopeMappings.ClientMappings[secondClientName] = secondClient
	return &scopeMappings, nil
}

// FAIL
func keycloakClientScopeMappingsTest() error {
	err := WaitForClientToBeReady(keycloakNamespace, testSecondKeycloakClientCRName)

	if err != nil {
		return err
	}
	keycloakCR, err := getDeployedKeycloakCR(keycloakNamespace)
	authenticatedClient, err := MakeAuthenticatedClient(*keycloakCR)
	if err != nil {
		return err
	}

	// create initial client with scope mappings
	client, err := getKeycloakClientWithScopeMappingsCR(
		authenticatedClient,
		[]string{"realmRoleA", "realmRoleB"},
		[]string{"a", "b"})
	if err != nil {
		return err
	}
	client, err = CreateKeycloakClient(client)
	fmt.Println(err)

	if err != nil {
		return err
	}
	err = WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)

	if err != nil {
		return err
	}

	// add non-existent roles
	var retrievedClient *keycloakv1alpha1.KeycloakClient
	retrievedClient, err = GetNamespacedKeycloakClient(keycloakNamespace, testKeycloakClientCRName)
	if err != nil {
		return err
	}
	GinkgoWriter.Print("add nonexisting role to %s", testKeycloakClientCRName)

	mappings, err := getKeycloakClientScopeMappings(
		authenticatedClient,
		[]string{"realmRoleB", "realmRoleC", "nonexistent"},
		[]string{"b", "c", "nonexistent"},
	)
	if err != nil {
		return err
	}
	retrievedClient.Spec.ScopeMappings = mappings
	GinkgoWriter.Print("update %s with nonexisting role", testKeycloakClientCRName)
	_, err = UpdateKeycloakClient(keycloakNamespace, retrievedClient)
	if err != nil {
		return err
	}
	GinkgoWriter.Print("wait for failing keycloakclient %s with nonexisting role", testKeycloakClientCRName)
	err = WaitForClientToBeFailing(keycloakNamespace, testKeycloakClientCRName)

	if err != nil {
		return fmt.Errorf("keycloakclient %s should be failing with nonexisting role but got %s", testKeycloakClientCRName, err)
	}

	// update client: delete/leave/create mappings
	retrievedClient, err = GetNamespacedKeycloakClient(keycloakNamespace, testKeycloakClientCRName)
	if err != nil {
		return err
	}
	mappings, err = getKeycloakClientScopeMappings(authenticatedClient, []string{"realmRoleB", "realmRoleC"}, []string{"b", "c"})
	if err != nil {
		return err
	}
	retrievedClient.Spec.ScopeMappings = mappings

	_, err = UpdateKeycloakClient(keycloakNamespace, retrievedClient)
	if err != nil {
		return err
	}
	err = WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)
	if err != nil {
		return err
	}

	retrievedMappings, err := authenticatedClient.ListScopeMappings(testKeycloakClientCRName, realmName)
	if err != nil {
		return err
	}
	expected := retrievedClient.Spec.ScopeMappings

	difference, intersection := model.RoleDifferenceIntersection(
		retrievedMappings.RealmMappings,
		expected.RealmMappings)
	Expect(0).To(Equal(len(difference)))
	Expect(2).To(Equal(len(intersection)))

	difference, intersection = model.RoleDifferenceIntersection(
		retrievedMappings.ClientMappings[secondClientName].Mappings,
		expected.ClientMappings[secondClientName].Mappings)
	Expect(0).To(Equal(len(difference)))
	Expect(2).To(Equal(len(intersection)))

	return nil
}

// FAIL
func keycloakClientServiceAccountRealmRolesTest() error {
	// deploy secondary client with a few client roles
	_, err := prepareKeycloakClientWithRolesCR()
	if err != nil {
		return err
	}
	err = WaitForClientToBeReady(keycloakNamespace, testSecondKeycloakClientCRName)
	if err != nil {
		return err
	}

	// deploy primary client with service account roles
	_, err = prepareKeycloakClientWithServiceAccount()
	if err != nil {
		return err
	}
	err = WaitForClientToBeReady(keycloakNamespace, testKeycloakClientCRName)
	if err != nil {
		return err
	}

	keycloakCR, err := getDeployedKeycloakCR(keycloakNamespace)

	// assert roles
	assertServiceAccountRoles(*keycloakCR, testKeycloakClientCRName, []string{"realmRoleA", "realmRoleB"}, map[string][]string{secondClientName: {"a", "b"}})

	// remove one of the roles
	var retrievedClient *keycloakv1alpha1.KeycloakClient
	retrievedClient, err = GetNamespacedKeycloakClient(keycloakNamespace, testKeycloakClientCRName)
	if err != nil {
		return err
	}
	retrievedClient.Spec.ServiceAccountRealmRoles = []string{"realmRoleB"}
	retrievedClient.Spec.ServiceAccountClientRoles = map[string][]string{secondClientName: {"b"}}
	_, err = UpdateKeycloakClient(keycloakNamespace, retrievedClient)
	if err != nil {
		return err
	}

	// assert roles were removed
	assertServiceAccountRoles(*keycloakCR, testKeycloakClientCRName, []string{"realmRoleB"}, map[string][]string{secondClientName: {"b"}})

	return nil
}

func assertServiceAccountRoles(keycloakCR keycloakv1alpha1.Keycloak, clientID string, expectedRealmRoles []string, expectedClientRoles map[string][]string) {
	err := WaitForConditionWithClient(keycloakCR, func(authenticatedClient common.KeycloakInterface) error {
		serviceAccountUser, err := authenticatedClient.GetServiceAccountUser(realmName, clientID)
		if err != nil {
			return err
		}

		// get realm role names
		actualRealmRoles, err := authenticatedClient.ListUserRealmRoles(realmName, serviceAccountUser.ID)
		if err != nil {
			return err
		}
		var actualRealmRolesNames []string
		for _, v := range actualRealmRoles {
			actualRealmRolesNames = append(actualRealmRolesNames, v.Name)
		}

		// get role names for all specified clients
		var actualClientRolesNames = map[string][]string{}
		for k := range expectedClientRoles {
			roles, err := authenticatedClient.ListUserClientRoles(realmName, k, serviceAccountUser.ID)
			if err != nil {
				return err
			}
			for _, v := range roles {
				actualClientRolesNames[k] = append(actualClientRolesNames[k], v.Name)
			}
		}

		// can't use standard asserts as it would fail the test; we want to fail only on timeout

		// sort arrays for proper comparison
		sort.Strings(expectedRealmRoles)
		sort.Strings(actualRealmRolesNames)
		for k := range expectedClientRoles {
			sort.Strings(expectedClientRoles[k])
		}
		for k := range actualClientRolesNames {
			sort.Strings(actualClientRolesNames[k])
		}

		if !reflect.DeepEqual(expectedRealmRoles, actualRealmRolesNames) {
			return errors.Errorf("Expected realm roles: %s, Actual: %s", expectedRealmRoles, actualRealmRolesNames)
		}

		// strings are the easiest way to compare maps
		if fmt.Sprint(expectedClientRoles) != fmt.Sprint(actualClientRolesNames) {
			return errors.Errorf("Expected client roles: %s, Actual: %s", expectedClientRoles, actualClientRolesNames)
		}

		return nil
	})
	Expect(err).To(BeNil())
}

func tearDownKeycloakClients() error {
	keycloakClientList, err := getKeycloakApiClient().KeycloakV1alpha1().KeycloakClients(keycloakNamespace).List(context.Background(), metav1.ListOptions{})

	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	for _, keycloakClient := range (*keycloakClientList).Items {
		name := keycloakClient.Name
		err = getKeycloakApiClient().KeycloakV1alpha1().KeycloakClients(keycloakNamespace).Delete(context.Background(), name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil

}
