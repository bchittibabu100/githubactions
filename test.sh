curl --header "X-Vault-Token:" --request GET http://172.21.67.14:8200/v1/sys/internal/ui/mounts/engineering/vpay360/vpay360-ui/dev/admin
{"request_id":"7fadd791-409f-47f4-d2d1-494294474bdf","lease_id":"","renewable":false,"lease_duration":0,"data":{"accessor":"kv_2ca5a707","config":{"default_lease_ttl":2764800,"force_no_cache":false,"max_lease_ttl":2764800},"deprecation_status":"supported","description":"Key Value secrets storage for VPAY Engineering","external_entropy_access":false,"local":false,"options":{"version":"2"},"path":"engineering/","plugin_version":"","running_plugin_version":"v0.19.0+builtin","running_sha256":"","seal_wrap":false,"type":"kv","uuid":"175c1286-4cb8-c80b-81ea-2bb1d6a1a0af"},"wrap_info":null,"warnings":null,"auth":null,"mount_type":""}


2024-10-21T18:31:40.7707662Z URL: GET http://172.21.67.14:8200/v1/sys/internal/ui/mounts/engineering/vpay360/vpay360-ui/dev/admin
2024-10-21T18:31:40.7709011Z Code: 403. Errors:
2024-10-21T18:31:40.7709494Z 
2024-10-21T18:31:40.7709809Z * permission denied
2024-10-21T18:31:40.7875295Z Error making API request.
