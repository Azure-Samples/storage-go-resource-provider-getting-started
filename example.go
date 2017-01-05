// This package demonstrates how to manage storage accounts with Azure Storage Resource Provider.
package main

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	location    = "westus"
	groupName   = "your-azure-sample-group"
	accountName = "golangrocksonazure"
	provider    = "Microsoft.Storage"
)

// This example requires that the following environment vars are set:
//
// AZURE_TENANT_ID: contains your Azure Active Directory tenant ID or domain
// AZURE_CLIENT_ID: contains your Azure Active Directory Application Client ID
// AZURE_CLIENT_SECRET: contains your Azure Active Directory Application Secret
// AZURE_SUBSCRIPTION_ID: contains your Azure Subscription ID
//

var (
	resourcesClient resources.ProvidersClient
	storageClient   storage.AccountsClient
	groupClient     resources.GroupsClient
	usageClient     storage.UsageOperationsClient
)

func init() {
	subscriptionID := getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	tenantID := getEnvVarOrExit("AZURE_TENANT_ID")

	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(tenantID)
	onErrorFail(err, "OAuthConfigForTenant failed")

	clientID := getEnvVarOrExit("AZURE_CLIENT_ID")
	clientSecret := getEnvVarOrExit("AZURE_CLIENT_SECRET")
	spToken, err := azure.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	onErrorFail(err, "NewServicePrincipalToken failed")

	createClients(subscriptionID, spToken)
}

func main() {
	registerResourceProvider()
	checkAccountAvailability()
	createResourceGroup()
	createStorageAccount()
	getStorageAccountProperties()
	listStorageAccountsByResourceGroup()
	listStorageAccountsBySubscription()
	keys := getStorageKeys()
	regenStorageKey(keys)
	updateStorageAccount()
	listUsage()

	fmt.Print(fmt.Sprintf("Press enter to delete the resource group '%s'... (y/n)", groupName))

	var input string
	fmt.Scanln(&input)
	if input == "y" {
		delete()
	}
}

// getEnvVarOrExit returns the value of specified environment variable or terminates if it's not defined.
func getEnvVarOrExit(varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		fmt.Printf("Missing environment variable %s\n", varName)
		os.Exit(1)
	}

	return value
}

func createClients(subscriptionID string, spToken *azure.ServicePrincipalToken) {
	resourcesClient = resources.NewProvidersClient(subscriptionID)
	resourcesClient.Authorizer = spToken

	storageClient = storage.NewAccountsClient(subscriptionID)
	storageClient.Authorizer = spToken

	groupClient = resources.NewGroupsClient(subscriptionID)
	groupClient.Authorizer = spToken

	usageClient = storage.NewUsageOperationsClient(subscriptionID)
	usageClient.Authorizer = spToken
}

func registerResourceProvider() {
	fmt.Println("Register resource provider...")
	_, err := resourcesClient.Register(provider)
	onErrorFail(err, "Register failed")
}

func checkAccountAvailability() {
	fmt.Println("Check account name availability...")
	result, err := storageClient.CheckNameAvailability(storage.AccountCheckNameAvailabilityParameters{
		Name: to.StringPtr(accountName),
		Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
	})
	onErrorFail(err, "CheckNameAvailability failed")

	if *result.NameAvailable == true {
		fmt.Printf("\t'%s' is available!\n", accountName)
	} else {
		fmt.Printf("\t'%s' is not available :(\n\tReason: %s\n\tMessage: %s\n",
			accountName,
			result.Reason,
			*result.Message)
		fmt.Println("No resources were created.")
		os.Exit(1)
	}
}

func createResourceGroup() {
	fmt.Println("Create resource group...")
	_, err := groupClient.CreateOrUpdate(groupName, resources.ResourceGroup{
		Location: to.StringPtr(location),
	})
	onErrorFail(err, "CreateOrUpdate failed")
}

func createStorageAccount() {
	fmt.Println("Create storage account...")
	_, err := storageClient.Create(groupName, accountName, storage.AccountCreateParameters{
		Sku: &storage.Sku{
			Name: storage.StandardLRS,
		},
		Location: to.StringPtr(location),
		AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
	}, nil)
	onErrorFail(err, "Create failed")
}

func getStorageAccountProperties() {
	fmt.Println("Get storage account properties...")
	account, err := storageClient.GetProperties(groupName, accountName)
	onErrorFail(err, "GetProperties failed")

	fmt.Printf("'%v' storage account properties\n", *account.Name)
	fmt.Printf("\t                ID: %v\n", *account.ID)
	fmt.Printf("\t          Sku Name: %v\n", account.Sku.Name)
	fmt.Printf("\t              Type: %v\n", *account.Type)
	fmt.Printf("\t          Location: %v\n", *account.Location)
	fmt.Printf("\tProvisioning State: %v\n", account.ProvisioningState)
}

func listStorageAccountsByResourceGroup() {
	fmt.Printf("List all storage accounts in '%s' resource group\n", groupName)
	list, err := storageClient.ListByResourceGroup(groupName)
	onErrorFail(err, "ListByResourceGroup failed")
	printAccountList(&list)
}

func listStorageAccountsBySubscription() {
	fmt.Println("List all storage accounts under the subscription")
	list, err := storageClient.List()
	onErrorFail(err, "List failed")
	printAccountList(&list)
}

func printAccountList(list *storage.AccountListResult) {
	for _, acc := range *list.Value {
		fmt.Printf("\t%s\n", *acc.Name)
	}
}

func getStorageKeys() *storage.AccountListKeysResult {
	fmt.Println("Get storage account keys...")
	keys, err := storageClient.ListKeys(groupName, accountName)
	onErrorFail(err, "ListKeys failed")

	fmt.Printf("'%s' storage account keys\n", accountName)
	for _, key := range *keys.Keys {
		fmt.Printf("\tKey name: %s\n\tValue: %s...\n\tPermissions: %s\n",
			*key.KeyName,
			(*key.Value)[:5],
			key.Permissions)
		fmt.Println("\t----------------")
	}

	return &keys
}

func regenStorageKey(keys *storage.AccountListKeysResult) {
	fmt.Println("Regenerate account key...")
	newKeys, err := storageClient.RegenerateKey(groupName, accountName, storage.AccountRegenerateKeyParameters{
		KeyName: (*keys.Keys)[0].KeyName},
	)
	onErrorFail(err, "RegenerateKey failed")

	fmt.Println("New key")
	fmt.Printf("\tKey name: %s\n\tValue: %s...\n\tPermissions: %s\n",
		*(*newKeys.Keys)[0].KeyName,
		(*(*newKeys.Keys)[0].Value)[:5],
		(*newKeys.Keys)[0].Permissions)
}

func updateStorageAccount() {
	fmt.Println("Update storage account...")
	_, err := storageClient.Update(groupName, accountName, storage.AccountUpdateParameters{
		Tags: &map[string]*string{
			"who rocks": to.StringPtr("golang"),
			"where":     to.StringPtr("on azure")},
	})
	onErrorFail(err, "Update failed")
}

func listUsage() {
	fmt.Println("List usage for storage accounts in subscription...")
	usageList, err := usageClient.List()
	onErrorFail(err, "List failed")

	for _, usage := range *usageList.Value {
		fmt.Printf("\t%v: %v / %v\n", *usage.Name.Value, *usage.CurrentValue, *usage.Limit)
	}
}

func delete() {
	fmt.Println("Delete storage account...")
	_, err := storageClient.Delete(groupName, accountName)
	onErrorFail(err, "Delete failed")

	fmt.Println("Delete resource group...")
	_, err = groupClient.Delete(groupName, nil)
	onErrorFail(err, "Delete failed")
}

// onErrorFail prints a failure message and exits the program if err is not nil.
func onErrorFail(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %s\n", message, err)
		os.Exit(1)
	}
}
