package client

import (
	"net"
	"testing"

	"github.com/octohelm/x/testing/bdd"
)

func TestHostAliases(t *testing.T) {
	bdd.FromT(t).Given("host alias for ipv4", func(b bdd.T) {
		ha := &HostAlias{
			IP: net.ParseIP("127.0.0.1"),
			Hostnames: []string{
				"localhost",
				"localhost1",
			},
		}

		txt := bdd.Must(ha.MarshalText())

		b.Then("match results",
			bdd.Equal("127.0.0.1:localhost,localhost1", string(txt)),
		)

		b.When("unmarshal", func(b bdd.T) {
			ha1 := &HostAlias{}

			b.Then("success",
				bdd.NoError(ha1.UnmarshalText(txt)),
				bdd.Equal(ha, ha1),
			)
		})
	})

	bdd.FromT(t).Given("host alias for ipv6", func(b bdd.T) {
		ha := &HostAlias{
			IP: net.ParseIP("::1"),
			Hostnames: []string{
				"localhost",
				"localhost1",
			},
		}

		txt := bdd.Must(ha.MarshalText())

		b.Then("match results",
			bdd.Equal("[::1]:localhost,localhost1", string(txt)),
		)

		b.When("unmarshal", func(b bdd.T) {
			ha1 := &HostAlias{}

			b.Then("success",
				bdd.NoError(ha1.UnmarshalText(txt)),
				bdd.Equal(ha, ha1),
			)
		})
	})
}
