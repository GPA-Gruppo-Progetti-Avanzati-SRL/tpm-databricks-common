# tpm-databricks-common

Go library for connecting to Databricks and executing SQL queries via the Statement Execution API.

## Configuration

The `Config` struct controls authentication and target warehouse. Available `auth-type` values:

| Value                 | Description                                | Implemented |
|-----------------------|--------------------------------------------|-------------|
| `azure-client-secret` | Azure service principal with client secret | true        |
| `azure-cli,azure-msi` | Azure CLI or managed identity              | false       |
| `azure-msi`           | Azure managed identity only                | false       |
| `oauth-m2m`           | Databricks OAuth machine-to-machine        | false       |
| `databricks-cli`      | Local Databricks CLI credentials           | false       |
| `google-id`           | Google identity token                      | false       |
| `google-credentials`  | Google service account credentials         | false       |

## Query resource patterns

SQL queries passed to `Find` may reference entries from `catalog-resources` using the pattern `[dbrks:<resource-id>]`. 
Before execution the pattern is replaced with the `name` of the matching resource. An unresolved pattern returns an error.

```sql
SELECT * FROM [dbrks:main].my_schema.my_table
-- resolved to →
SELECT * FROM Main Catalog.my_schema.my_table
```

The `id` is the short alias used inside queries; `name` is the actual Databricks catalog or schema identifier.

### Azure — service principal (client secret)

```json
{
  "name": "my-databricks-connection",
  "host": "https://<workspace-id>.azuredatabricks.net",
  "workspace-resource-id": "/subscriptions/<sub-id>/resourceGroups/<rg>/providers/Microsoft.Databricks/workspaces/<ws>",
  "auth-type": "azure-client-secret",
  "service-principal": {
    "tenant-id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "client-id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "client-secret": "your-client-secret"
  },
  "warehouse-id": "xxxxxxxxxxxx",
  "catalog-resources": [
    { "id": "main", "name": "Main Catalog" },
    { "id": "hive_metastore", "name": "Legacy Hive" }
  ]
}
```

