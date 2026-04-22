package client

import (
	"net"
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestHostAliasMarshalAndUnmarshal(t *testing.T) {
	t.Run("ipv4 host alias", func(t *testing.T) {
		ha := &HostAlias{
			IP: net.ParseIP("127.0.0.1"),
			Hostnames: []string{
				"localhost",
				"localhost1",
			},
		}

		Then(t, "序列化与反序列化结果正确",
			ExpectMust(func() error {
				txt, err := ha.MarshalText()
				if err != nil {
					return err
				}
				if string(txt) != "127.0.0.1:localhost,localhost1" {
					return errClient("unexpected ipv4 marshal result")
				}

				ha1 := &HostAlias{}
				if err := ha1.UnmarshalText(txt); err != nil {
					return err
				}
				if !reflect.DeepEqual(ha, ha1) {
					return errClient("unexpected ipv4 unmarshal result")
				}
				return nil
			}),
		)
	})

	t.Run("ipv6 host alias", func(t *testing.T) {
		ha := &HostAlias{
			IP: net.ParseIP("::1"),
			Hostnames: []string{
				"localhost",
				"localhost1",
			},
		}

		Then(t, "序列化与反序列化结果正确", ExpectMust(func() error {
			txt, err := ha.MarshalText()
			if err != nil {
				return err
			}
			if string(txt) != "[::1]:localhost,localhost1" {
				return errClient("unexpected ipv6 marshal result")
			}
			ha1 := &HostAlias{}
			if err := ha1.UnmarshalText(txt); err != nil {
				return err
			}
			if !reflect.DeepEqual(ha, ha1) {
				return errClient("unexpected ipv6 unmarshal result")
			}
			return nil
		}))
	})
}
