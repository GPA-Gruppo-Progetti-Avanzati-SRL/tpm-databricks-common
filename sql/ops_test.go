package sql_test

import (
	"context"
	"os"
	"testing"

	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-databricks-common/dbricksLks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-databricks-common/sql"
	"github.com/stretchr/testify/require"
)

func TestOps(t *testing.T) {
	lks, err := dbricksLks.Initialize([]dbricksLks.Config{
		{
			Name:                "default",
			Host:                "https://adb-6686371009486131.11.azuredatabricks.net",
			WorkspaceResourceID: "/subscriptions/9a9b2fd9-6706-42f0-9131-5a726dffcb94/resourceGroups/SALES-INFRA-COMMON-SVIL/providers/Microsoft.Databricks/workspaces/ssalescomdatab01azne",
			AuthType:            dbricksLks.AuthTypeAzureClientSecret,
			ServicePrincipal: &dbricksLks.ServicePrincipal{
				TenantID:     os.Getenv("AZURE_TENANT_ID"),
				ClientID:     os.Getenv("AZURE_CLIENT_ID"),
				ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
			},
			WarehouseID: "2cb583d2b287b9e7",
			Resources: []dbricksLks.Resource{
				{
					Id:   "smart-catalog",
					Name: "sales_dbcatalog_ssalescomdatab01azne.smartcatalog_trusted_dss",
				},
			},
		},
	})
	require.NoError(t, err)

	opr, b, err := sql.JsonFind(context.Background(), lks[0], "SHOW TABLES IN [dbrks:smart-catalog]")
	require.NoError(t, err)
	t.Log(opr, string(b))

	opr, b, err = sql.JsonFindOne(context.Background(), lks[0], "SHOW TABLES IN [dbrks:smart-catalog]")
	require.NoError(t, err)
	t.Log(opr, string(b))

}
