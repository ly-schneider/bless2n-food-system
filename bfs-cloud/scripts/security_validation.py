#!/usr/bin/env python3

import json
import sys
from azure.identity import DefaultAzureCredential
from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.security import SecurityCenter
from azure.mgmt.network import NetworkManagementClient

def check_security_configuration(subscription_id, resource_group_name, environment):
    """
    Validate security configuration against OWASP Cloud Security principles
    """
    
    credential = DefaultAzureCredential()
    security_client = SecurityCenter(credential, subscription_id)
    network_client = NetworkManagementClient(credential, subscription_id)
    resource_client = ResourceManagementClient(credential, subscription_id)
    
    security_results = {
        "owasp_compliance": {
            "identity_and_access": {"status": "checking", "score": 0},
            "logging_and_monitoring": {"status": "checking", "score": 0},
            "data_protection": {"status": "checking", "score": 0},
            "network_security": {"status": "checking", "score": 0},
            "secure_configuration": {"status": "checking", "score": 0}
        },
        "security_alerts": [],
        "recommendations": []
    }
    
    try:
        # Check Network Security Groups
        nsgs = list(network_client.network_security_groups.list(resource_group_name))
        if nsgs:
            security_results["owasp_compliance"]["network_security"]["status"] = "compliant"
            security_results["owasp_compliance"]["network_security"]["score"] = 90
        else:
            security_results["owasp_compliance"]["network_security"]["status"] = "non_compliant"
            security_results["owasp_compliance"]["network_security"]["score"] = 30
            security_results["recommendations"].append(
                "Implement Network Security Groups for network segmentation"
            )
        
        # Check Security Center assessments
        assessments = list(security_client.assessments.list(
            scope=f"/subscriptions/{subscription_id}/resourceGroups/{resource_group_name}"
        ))
        
        high_severity_issues = [a for a in assessments if a.status.severity == "High"]
        if len(high_severity_issues) == 0:
            security_results["owasp_compliance"]["secure_configuration"]["status"] = "compliant"
            security_results["owasp_compliance"]["secure_configuration"]["score"] = 95
        else:
            security_results["owasp_compliance"]["secure_configuration"]["status"] = "non_compliant"
            security_results["owasp_compliance"]["secure_configuration"]["score"] = 50
            security_results["security_alerts"].extend([
                f"High severity issue: {issue.display_name}" for issue in high_severity_issues[:5]
            ])
        
        # Check for Key Vault usage
        resources = list(resource_client.resources.list_by_resource_group(resource_group_name))
        key_vaults = [r for r in resources if r.type == "Microsoft.KeyVault/vaults"]
        
        if key_vaults:
            security_results["owasp_compliance"]["data_protection"]["status"] = "compliant"
            security_results["owasp_compliance"]["data_protection"]["score"] = 85
        else:
            security_results["owasp_compliance"]["data_protection"]["status"] = "non_compliant"
            security_results["owasp_compliance"]["data_protection"]["score"] = 40
            security_results["recommendations"].append(
                "Implement Azure Key Vault for secrets management"
            )
        
        # Check for Log Analytics workspace
        log_workspaces = [r for r in resources if r.type == "Microsoft.OperationalInsights/workspaces"]
        if log_workspaces:
            security_results["owasp_compliance"]["logging_and_monitoring"]["status"] = "compliant"
            security_results["owasp_compliance"]["logging_and_monitoring"]["score"] = 88
        else:
            security_results["owasp_compliance"]["logging_and_monitoring"]["status"] = "non_compliant"
            security_results["owasp_compliance"]["logging_and_monitoring"]["score"] = 35
            security_results["recommendations"].append(
                "Implement centralized logging with Log Analytics"
            )
        
        # Check managed identities
        container_apps = [r for r in resources if r.type == "Microsoft.App/containerApps"]
        identity_score = 70 if container_apps else 50
        security_results["owasp_compliance"]["identity_and_access"]["status"] = "compliant" if identity_score > 60 else "non_compliant"
        security_results["owasp_compliance"]["identity_and_access"]["score"] = identity_score
        
        # Calculate overall compliance score
        total_score = sum([
            security_results["owasp_compliance"][category]["score"] 
            for category in security_results["owasp_compliance"]
        ]) / len(security_results["owasp_compliance"])
        
        security_results["overall_score"] = int(total_score)
        security_results["compliance_level"] = (
            "high" if total_score >= 80 else
            "medium" if total_score >= 60 else
            "low"
        )
        
    except Exception as e:
        security_results["error"] = str(e)
        security_results["overall_score"] = 0
        security_results["compliance_level"] = "unknown"
    
    return security_results

def main():
    # Read input from Terraform
    input_data = json.loads(sys.stdin.read())
    
    subscription_id = input_data.get("subscription_id", "")
    resource_group_name = input_data.get("resource_group_name", "")
    environment = input_data.get("environment", "")
    
    if not all([subscription_id, resource_group_name]):
        print(json.dumps({"error": "Missing required parameters"}))
        sys.exit(1)
    
    result = check_security_configuration(subscription_id, resource_group_name, environment)
    
    print(json.dumps(result))

if __name__ == "__main__":
    main()