package scvmm

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/masterzen/winrm"
)

func testBasicPreCheckVMStop(t *testing.T) {

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

func TestAccVmstop_Basic(t *testing.T) {
	vmName := "TestSujay"
	vmmServer := "WIN-2F929KU8HIU"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVmstopDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckVMStopConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVMStopExists("scvmm_stop_vm.StopVM", vmName, vmmServer),
					resource.TestCheckResourceAttr(
						"scvmm_stop_vm.StopVM", "vm_name", "TestSujay"),
				),
			},
		},
	})
}

func testAccCheckVmstopDestroy(s *terraform.State) error {
	return nil
}

func testAccCheckVMStopExists(n, vmName string, vmmServer string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Vm ID is set")
		}

		org := testAccProvider.Meta().(*winrm.Client)
		script := "[CmdletBinding(SupportsShouldProcess=$true)]\nparam (\n\n  [Parameter(Mandatory=$true,HelpMessage=\"Enter VM Name\")]\n  [string]$vmName,\n\n  [Parameter(Mandatory=$true,HelpMessage=\"Enter VmmServer\")]\n  [string]$vmmServer\n\n)\n\nBegin\n{\n   \n            If (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] \"Administrator\"))\n          {   \n                $arguments = \"\" + $myinvocation.mycommand.definition + \" \"\n                $myinvocation.BoundParameters.Values | foreach{$arguments += \"'$_' \" }\n            echo $arguments\n            Start-Process powershell -Verb runAs -ArgumentList $arguments\n            Break\n         }\n\u0009    try\n\u0009     {    \n                $VMs = Get-SCVirtualMachine -VMMServer $vmmServer  | where-Object { $_.Name -Match $vmName -And $_.Status -eq \"PowerOff\" }               \n                if($VMs -eq $null)\n                {     \n                  Write-Error \"VM is not Stopped \"        \n                 }  \n                              \n            }\n\u0009     catch [Exception]\n\u0009       { echo $_.Exception.Message\n\u0009        }\n}\n"
		arguments := vmName + " " + vmmServer
		filename := "stopvm_test"
		result, err := execScript(org, script, filename, arguments)

		if err != "" {
			return fmt.Errorf("Error , VM is not stopped %v", result)
		}

		return nil
	}
}

const testAccCheckVMStopConfigBasic = `
resource "scvmm_stop_vm" "StopVM"{
        vm_name = "TestSujay"
		timeout= "1000"
        vmm_server = "WIN-2F929KU8HIU"
		    
       
}`