# githubactions
github action workflows to enable CI/CD


vault helper go:
################
package helpers

import (
	"context"
	"strings"
	"sync"

	vault "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

type VaultHelper struct {
	Client *vault.Client
}

func CreateVaultHelper(c *vault.Config) (*VaultHelper, error) {
	client, err := vault.NewClient(c)
	if err != nil {
		return nil, err
	}
	return &VaultHelper{Client: client}, nil
}

// Deprecated: Use VaultHelper.ReadSecretSimple
func (v *VaultHelper) ReadSecret(path string) (*vault.Secret, bool, error) {
	mountPath, v2, err := IsKVv2(path, v.Client)
	if err != nil {
		return nil, v2, err
	}
	if v2 {
		path = AddPrefixToVKVPath(path, mountPath, "data")
	}
	secret, err := v.Client.Logical().Read(path)
	if err != nil {
		return nil, v2, err
	}
	if secret == nil {
		return nil, v2, errors.Errorf("path: %s: Secret not found at path", path)
	}
	return secret, v2, nil
}

func (v *VaultHelper) ReadData(path string) (map[string]interface{}, error) {
	secret, v2, err := v.ReadSecret(path)
	if err != nil {
		return nil, err
	}
	data := secret.Data
	if v2 && data != nil {
		data = nil
		dataRaw := secret.Data["data"]

		if dataRaw != nil {
			data = dataRaw.(map[string]interface{})
		}
	}
	return data, nil
}

/*
Uses new vault api features to read secrets and keys rather than spinning our own
*/
func (v *VaultHelper) ReadKeySimple(path string, key string) (string, error) {
	secret, err := v.ReadSecretSimple(path)
	if err != nil {
		return "", err
	}
	data, ok := secret.Data[key]
	if !ok {
		return "", errors.Errorf("path: %s, key: %s: No value found in secret", path, key)
	}
	return data.(string), nil
}

/*
Uses new vault api features to read secrets rather than spinning our own
*/
func (v *VaultHelper) ReadSecretSimple(path string) (*vault.KVSecret, error) {
	context := context.Background()
	mountPath, v2, err := IsKVv2(path, v.Client)
	p := strings.TrimPrefix(path, mountPath)
	if err != nil {
		return nil, err
	}
	if v2 {
		client := v.Client.KVv2(mountPath)
		return client.Get(context, p)
	} else {
		client := v.Client.KVv1(mountPath)
		return client.Get(context, p)
	}
}

/*
Part of the logic in this method is to help out Logical read kv v2 secrets, there is an
open issue for this to move this logic into Logical https://github.com/hashicorp/vault/issues/11853
*/
// Deprecated: Use VaultHelper.ReadKeySimple
func (v *VaultHelper) ReadKey(path string, key string) (string, error) {
	data, err := v.ReadData(path)
	if err != nil {
		return "", err
	}

	if data[key] == nil {
		return "", errors.Errorf("path: %s, key: %s: No value found in secret", path, key)
	}
	secretString := data[key].(string)

	return secretString, nil

}

func (v *VaultHelper) List(path string) ([]interface{}, error) {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	mountPath, v2, err := IsKVv2(path, v.Client)
	if err != nil {
		return nil, err
	}

	if v2 {
		path = AddPrefixToVKVPath(path, mountPath, "metadata")
		if err != nil {
			return nil, err
		}
	}

	secret, err := v.Client.Logical().List(path)
	if err != nil {
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return nil, errors.Errorf("path: %s: No secret found at path", path)
	}
	k, ok := secret.Data["keys"]
	if !ok {
		return nil, errors.Errorf("path: %s: No keys found in secret", path)
	}
	i, ok := k.([]interface{})
	if !ok {
		return nil, errors.Errorf("path: %s: Keys are not an array, possible unsupported backend", path)
	}
	return i, nil
}

func (v *VaultHelper) Write(path string, data map[string]interface{}, force bool) (*vault.Secret, error) {
	mountPath, v2, err := IsKVv2(path, v.Client)
	if err != nil {
		return nil, err
	}
	if v2 {
		path = AddPrefixToVKVPath(path, mountPath, "data")
		data = map[string]interface{}{
			"data":    data,
			"options": map[string]interface{}{},
		}
		if !force {
			data["options"].(map[string]interface{})["cas"] = 0
		}
	}

	secret, err := v.Client.Logical().Write(path, data)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

type VaultSecretOperation func(mountPath, secretPath string)

func (v *VaultHelper) OperateOnSecrets(mountPath string, operation VaultSecretOperation) error {
	mountPath = strings.TrimSuffix(mountPath, "/") + "/"
	wg := sync.WaitGroup{}
	err := v.recurseSecrets(&wg, mountPath, "", operation)
	if err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func (v *VaultHelper) recurseSecrets(wg *sync.WaitGroup, mountPath string, secretPath string, operation VaultSecretOperation) error {
	secret, err := v.Client.Logical().List(mountPath + "metadata/" + secretPath)
	if err != nil {
		return err
	}
	for _, item := range secret.Data["keys"].([]interface{}) {
		path := item.(string)
		if strings.HasSuffix(path, "/") {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v.recurseSecrets(wg, mountPath, secretPath+path, operation)
			}()
		} else {
			operation(mountPath, secretPath+path)
		}
	}
	return nil
}


kv helper go:
#############
