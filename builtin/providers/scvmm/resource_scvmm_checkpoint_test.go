package scvmm

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/masterzen/winrm"
)

func testBasicPreCheckSP(t *testing.T) {

	testAccPreCheck(t)

	if v := os.Getenv("SCVMM_SERVER_IP"); v == "" {
		t.Fatal("SCVMM_SERVER_IP must be set for acceptance tests")
	}

	if v := os.Getenv("SCVMM_SERVER_PORT"); v == "" {
		t.Fatal("SCVMM_SERVER_PORT must be set for acceptance tests")
	}

	if v := os.Getenv("SCVMM_SERVER_USER"); v == "" {
		t.Fatal("SCVMM_SERVER_USER must be set for acceptance tests")
	}

	if v := os.Getenv("SCVMM_SERVER_PASSWORD"); v == "" {
		t.Fatal("SCVMM_SERVER_PASSWORD must be set for acceptance tests")
	}
}

func TestAccsp_Basic(t *testing.T) {
	vmName := "TestSujay"
	checkpointName := "testcheckpoint"
	vmmServer := "WIN-2F929KU8HIU"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCheckpointDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckCheckpointConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCheckpointExists("scvmm_checkpoint.CreateCheckpoint", vmName, vmmServer, checkpointName),
					resource.TestCheckResourceAttr(
						"scvmm_checkpoint.CreateCheckpoint", "vm_name", "TestSujay"),
					resource.TestCheckResourceAttr(
						"scvmm_checkpoint.CreateCheckpoint", "vmm_server", "WIN-2F929KU8HIU"),
					resource.TestCheckResourceAttr(
						"scvmm_checkpoint.CreateCheckpoint", "checkpoint_name", "testcheckpoint"),
				),
			},
		},
	})
}

func testAccCheckCheckpointDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {

		if rs.Type != "scvmm_checkpoint" {
			continue
		}
		org := testAccProvider.Meta().(*winrm.Client)

		script := "[CmdletBinding(SupportsShouldProcess=$true)]\nparam(\n    [parameter(Mandatory=$true,HelpMessage=\"Enter VMMServer\")]\n    [string]$vmmServer,\n\n    [parameter(Mandatory=$true,HelpMessage=\"Enter Virtual Machine Name\")]\n    [string]$vmName,\n\n    [parameter(Mandatory=$true,HelpMessage=\"Enter Checkpoint Name\")]\n    [string]$checkpointName\n)\nBegin\n{  \n       If (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] \"Administrator\"))\n    {   $arguments = \"\" + $myinvocation.mycommand.definition + \" \"\n        $myinvocation.BoundParameters.Values | foreach{\n            $arguments += \"'$_' \"\n        }\n        echo $arguments\n        Start-Process powershell -Verb runAs -ArgumentList $arguments\n        Break\n    }                 \n    \n        try\n\u0009     {\n\u0009\u0009 Set-SCVMMServer -VMMServer $vmmServer > $null\n                 \n                 $checkpoint = Get-SCVMCheckpoint -VM $vmName | Where-Object {$_.Name -eq $checkpointName}\n                 if($checkpoint-eq $null)\n                  {\n                    Write-Error \"No Checkpoint found\"\n                  }             \n             }catch [Exception]\n\u0009        {\n\u0009\u0009        echo $_.Exception.Message\n                }    \n}"
		arguments := rs.Primary.Attributes["vmm_server"] + " " + rs.Primary.Attributes["vm_name"] + " " + rs.Primary.Attributes["checkpoint_name"]
		filename := "deletecp"
		result, err := execScript(org, script, filename, arguments)

		if err == "" {
			return fmt.Errorf("Checkpoint  still exists: %v", result)
		}
	}

	return nil
}

func testAccCheckCheckpointExists(n, vmName string, vmmServer string, checkpointName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Vm ID is set")
		}

		org := testAccProvider.Meta().(*winrm.Client)

		script := "[CmdletBinding(SupportsShouldProcess=$true)]\nparam(\n    [parameter(Mandatory=$true,HelpMessage=\"Enter VMMServer\")]\n    [string]$vmmServer,\n\n    [parameter(Mandatory=$true,HelpMessage=\"Enter Virtual Machine Name\")]\n    [string]$vmName,\n\n    [parameter(Mandatory=$true,HelpMessage=\"Enter Checkpoint Name\")]\n    [string]$checkpointName\n)\nBegin\n{  \n       If (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] \"Administrator\"))\n    {   $arguments = \"\" + $myinvocation.mycommand.definition + \" \"\n        $myinvocation.BoundParameters.Values | foreach{\n            $arguments += \"'$_' \"\n        }\n        echo $arguments\n        Start-Process powershell -Verb runAs -ArgumentList $arguments\n        Break\n    }                 \n    \n        try\n\u0009     {\n\u0009\u0009 Set-SCVMMServer -VMMServer $vmmServer > $null\n                 \n                 $checkpoint = Get-SCVMCheckpoint -VM $vmName | Where-Object {$_.Name -eq $checkpointName}\n                 if($checkpoint-eq $null)\n                  {\n                    Write-Error \"No Checkpoint found\"\n                  }             \n             }catch [Exception]\n\u0009        {\n\u0009\u0009        echo $_.Exception.Message\n                }    \n}"
		arguments := vmmServer + " " + vmName + " " + checkpointName
		filename := "createcp"
		result, err := execScript(org, script, filename, arguments)

		if err != "" {
			return fmt.Errorf("Error while getting the checkpoint %v", result)
		}

		return nil
	}
}

const testAccCheckCheckpointConfigBasic = `
resource "scvmm_checkpoint" "CreateCheckpoint"{
     	timeout = "1000"
        vmm_server = "WIN-2F929KU8HIU"
        vm_name = "TestSujay"
        checkpoint_name= "testcheckpoint"
}`
