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
	westUS      = "westus"
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
	subscriptionID string
	spToken        *azure.ServicePrincipalToken
)

func init() {
	subscriptionID = getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	tenantID := getEnvVarOrExit("AZURE_TENANT_ID")

	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(tenantID)
	onErrorFail(err, "OAuthConfigForTenant failed")

	clientID := getEnvVarOrExit("AZURE_CLIENT_ID")
	clientSecret := getEnvVarOrExit("AZURE_CLIENT_SECRET")
	spToken, err = azure.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	onErrorFail(err, "NewServicePrincipalToken failed")
}

func main() {
	fmt.Println("Register resource provider...")
	resourcesClient := resources.NewProvidersClient(subscriptionID)
	resourcesClient.Authorizer = spToken

	_, err := resourcesClient.Register(provider)
	onErrorFail(err, "Register failed")

	fmt.Println("Check account name availability...")
	storageClient := storage.NewAccountsClient(subscriptionID)
	storageClient.Authorizer = spToken

	result, err := storageClient.CheckNameAvailability(storage.AccountCheckNameAvailabilityParameters{
		Name: to.StringPtr(accountName),
		Type: to.StringPtr("Microsoft.Storage/storageAccounts")},
	)
	onErrorFail(err, "CheckNameAvailability failed")

	if *result.NameAvailable == true {
		fmt.Printf("\t'%s' is available!\n", accountName)
	} else {
		fmt.Printf("\t'%s' is not available :(\n\tReason: %s\n\tMessage: %s\n",
			accountName,
			result.Reason,
			*result.Message)
	}

	fmt.Println("Create resource group...")
	groupClient := resources.NewGroupsClient(subscriptionID)
	groupClient.Authorizer = spToken

	_, err = groupClient.CreateOrUpdate(groupName, resources.ResourceGroup{
		Location: to.StringPtr(westUS)},
	)
	onErrorFail(err, "CreateOrUpdate failed")

	fmt.Println("Create storage account...")
	_, err = storageClient.Create(groupName, accountName, storage.AccountCreateParameters{
		Sku: &storage.Sku{
			Name: storage.StandardLRS},
		Location: to.StringPtr(westUS),
		AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
	}, nil)
	onErrorFail(err, "Create failed")

	fmt.Println("Get storage account properties...")
	account, err := storageClient.GetProperties(groupName, accountName)
	onErrorFail(err, "GetProperties failed")

	fmt.Printf("'%v' storage account properties\n", *account.Name)
	fmt.Printf("\t                ID: %v\n", *account.ID)
	fmt.Printf("\t          Sku Name: %v\n", account.Sku.Name)
	fmt.Printf("\t              Type: %v\n", *account.Type)
	fmt.Printf("\t          Location: %v\n", *account.Location)
	fmt.Printf("\tProvisioning State: %v\n", account.ProvisioningState)

	fmt.Printf("List all storage accounts in '%s' resource group\n", groupName)
	listGroupAccounts, err := storageClient.ListByResourceGroup(groupName)
	onErrorFail(err, "ListByResourceGroup failed")

	for _, acc := range *listGroupAccounts.Value {
		fmt.Printf("\t%s\n", *acc.Name)
	}

	fmt.Println("List all storage accounts under the subscription")
	listSubAccounts, err := storageClient.List()
	onErrorFail(err, "List failed")

	for _, acc := range *listSubAccounts.Value {
		fmt.Printf("\t%s\n", *acc.Name)
	}

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

	fmt.Println("Regenerate account key...")
	keys, err = storageClient.RegenerateKey(groupName, accountName, storage.AccountRegenerateKeyParameters{
		KeyName: (*keys.Keys)[0].KeyName},
	)
	onErrorFail(err, "RegenerateKey failed")

	fmt.Println("New key")
	fmt.Printf("\tKey name: %s\n\tValue: %s...\n\tPermissions: %s\n",
		*(*keys.Keys)[0].KeyName,
		(*(*keys.Keys)[0].Value)[:5],
		(*keys.Keys)[0].Permissions)

	fmt.Println("Update storage account...")
	_, err = storageClient.Update(groupName, accountName, storage.AccountUpdateParameters{
		Tags: &map[string]*string{
			"who rocks": to.StringPtr("golang"),
			"where":     to.StringPtr("on azure")},
	})
	onErrorFail(err, "Update failed")

	fmt.Println("List usage for storage accounts in subscription...")
	usageClient := storage.NewUsageOperationsClient(subscriptionID)
	usageClient.Authorizer = spToken

	usageList, err := usageClient.List()
	onErrorFail(err, "List failed")

	for _, usage := range *usageList.Value {
		fmt.Printf("\t%v: %v / %v\n", *usage.Name.Value, *usage.CurrentValue, *usage.Limit)
	}

	fmt.Print("Press enter to delete the storage account...")

	var input string
	fmt.Scanln(&input)

	fmt.Println("Delete storage account...")
	_, err = storageClient.Delete(groupName, accountName)
	onErrorFail(err, "Delete failed")

	fmt.Println("Delete resource group...")
	_, err = groupClient.Delete(groupName, nil)
	onErrorFail(err, "Delete failed")
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

// onErrorFail prints a failure message and exits the program if err is not nil.
func onErrorFail(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %s", message, err)
		os.Exit(1)
	}
}
