package personio

import (
	"errors"
	"fmt"

	"github.com/MarkusFreitag/keepassxc-go/pkg/keepassxc"
	"github.com/MarkusFreitag/keepassxc-go/pkg/keystore"
)

func fetchKeepassCredentials(url string) (user, pw, twoFactor string, err error) {
	client, err := createClient()
	if err != nil {
		return "", "", "", err
	}
	defer client.Disconnect()

	entries, err := client.GetLogins(url)
	if err != nil {
		return "", "", "", err
	}
	if len(entries) == 0 {
		return "", "", "", errors.New("no KeePassXC entries found")
	}
	if len(entries) > 1 {
		return "", "", "", errors.New("multiple KeePassXC entries found")
	}
	entry := entries[0]
	user = entry.Login
	pw = entry.Password.Plaintext()
	twoFactor, err = client.GetTOTP(entry.UUID)
	if err != nil {
		return "", "", "", fmt.Errorf("could not get TOTP from KeePassXC: %w", err)
	}

	return user, pw, twoFactor, nil
}

func createClient() (*keepassxc.Client, error) {
	store, err := keystore.Load()
	if err != nil {
		return nil, err
	}

	profile, err := store.DefaultProfile()
	if err != nil {
		return nil, err
	}

	keepassSocket, err := keepassxc.SocketPath()
	if err != nil {
		return nil, err
	}
	client := keepassxc.NewClient(keepassSocket, profile.Name, profile.NaclKey(), keepassxc.WithApplicationName(applicationName))
	if err := client.Connect(); err != nil {
		return nil, err
	}

	if err := client.ChangePublicKeys(); err != nil {
		return nil, err
	}

	if key := profile.NaclKey(); key == nil {
		if err := client.Associate(); err != nil {
			return nil, err
		}

		profile.Name, profile.Key = client.GetAssociatedProfile()

		if err := store.Add(profile); err != nil {
			return nil, err
		}

		if err := store.Save(); err != nil {
			return nil, err
		}
	} else {
		if err := client.TestAssociate(); err != nil {
			return nil, err
		}
	}

	return client, nil
}

const applicationName = "rootless-personio"
