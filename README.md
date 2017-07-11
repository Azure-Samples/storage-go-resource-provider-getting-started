---
services: storage
platforms: go
author: mcardosos
---

# Getting Started with Azure Storage Resource Provider in Go

This package demonstrates how to manage storage accounts with Azure Storage Resource Provider. The Storage Resource Provider is a client library for working with the storage accounts in your Azure subscription. Using the client library, you can create a new storage account, read its properties, list all storage accounts in a given subscription or resource group, read and regenerate the storage account keys, and delete a storage account.

If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](https://azure.microsoft.com/pricing/free-trial).

**On this page**

- [Run this sample](#run)
- [What does example.go do?](#sample)
    - [Register the Storage Resource Provider](#register)
    - [Check storage account name availability](#check)
    - [Create a new storage account](#createsa)
    - [Get the properties of a storage account](#get)
    - [List storage accounts by resource group](#listsarg)
    - [List storage accounts in subscription](#listsasyb)
    - [Get the storage account keys](#getkeys)
    - [Regenerate a storage account key](#regenkey)
    - [Update the storage account](#update)
    - [List usage](#listusage)
    - [Delete storage account](#delete)
- [More information](#info)

<a id="run"></a>

## Run this sample

1. If you don't already have it, [install Go](https://golang.org/dl/).

1. Clone the repository.

```
git clone https://github.com:Azure-Samples/app-service-web-go-manage.git
```

1. Install the dependencies using glide.

```
cd app-service-web-go-manage
glide install
```

1. Create an Azure service principal either through
    [Azure CLI](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal-cli/),
    [PowerShell](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal/)
    or [the portal](https://azure.microsoft.com/documentation/articles/resource-group-create-service-principal-portal/).

1. Set the following environment variables using the information from the service principle that you created.

```
export AZURE_TENANT_ID={your tenant id}
export AZURE_CLIENT_ID={your client id}
export AZURE_CLIENT_SECRET={your client secret}
export AZURE_SUBSCRIPTION_ID={your subscription id}
```

    > [AZURE.NOTE] On Windows, use `set` instead of `export`.

1. Run the sample.

```
go run example.go
```

<a id="sample"></a>

## What does example.go do?

The sample checks storage account name availability, creates a new storage account, gets the storage account properties, lists the storage accounts in the subscription or resource group, lists the storage account keys, regenerates the storage account keys, updates the storage account SKU, and deletes the storage account.

<a id="register"></a>

### Register the Storage Resource Provider

```go
resourcesClient.Register(provider)
```

<a id="check"></a>

### Check storage account name availability

Check the validity and availability of a given string as a storage account.

```go
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
```

A storage account needs a resource group to be created.

```go
groupClient.CreateOrUpdate(groupName, resources.Group{
    Location: to.StringPtr(location),
})
```

<a id="createsa"></a>

### Create a new storage account

```go
storageClient.Create(groupName, accountName, storage.AccountCreateParameters{
    Sku: &storage.Sku{
        Name: storage.StandardLRS,
    },
    Location: to.StringPtr(location),
    AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
}, nil)
```

<a id="get"></a>

### Get the properties of a storage account

```go
account, err := storageClient.GetProperties(groupName, accountName)
```

<a id="listsarg"></a>

### List storage accounts by resource group

```go
listGroupAccounts, err := storageClient.ListByResourceGroup(groupName)
onErrorFail(err, "ListByResourceGroup failed")

for _, acc := range *listGroupAccounts.Value {
     fmt.Printf("\t%s\n", *acc.Name)
}
```

<a id="listsasub"></a>

### List storage accounts in subscription

```go
listSubAccounts, err := storageClient.List()
onErrorFail(err, "List failed")

for _, acc := range *listSubAccounts.Value {
    fmt.Printf("\t%s\n", *acc.Name)
}
```

<a id="getkeys"></a>

### Get the storage account keys

```go
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
```

<a id="regenkey"></a>

### Regenerate a storage account key

```go
keys, err = storageClient.RegenerateKey(groupName, accountName, storage.AccountRegenerateKeyParameters{
    KeyName: (*keys.Keys)[0].KeyName},
)
```

<a id="update"></a>

### Update the storage account

Just like all resources, storage accounts can be updated.

```go
storageClient.Update(groupName, accountName, storage.AccountUpdateParameters{
    Tags: &map[string]*string{
        "who rocks": to.StringPtr("golang"),
        "where":     to.StringPtr("on azure")},
})
```

<a id="listusage"></a>

### List usage

```go
usageList, err := usageClient.List()
onErrorFail(err, "List failed")

for _, usage := range *usageList.Value {
    fmt.Printf("\t%v: %v / %v\n", *usage.Name.Value, *usage.CurrentValue, *usage.Limit)
}
```

<a id="delete"></a>

### Delete storage account

```go
storageClient.Delete(groupName, accountName)
```

At this point, the sample also deletes the resource group that it created.

```go
groupsClient.Delete(groupName, nil)
```

<a id="info"></a>

## More information

Please refer to [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) for more information.
***

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.