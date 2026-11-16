package compassconnection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractSubject(t *testing.T) {
	t.Run("should parse well-formed subject", func(t *testing.T) {
		result := extractSubject("O=Org,OU=OrgUnit,L=locality,ST=province,C=DE,CN=test")
		assert.Equal(t, "test", result["CN"])
		assert.Equal(t, "Org", result["O"])
		assert.Equal(t, "OrgUnit", result["OU"])
		assert.Equal(t, "locality", result["L"])
		assert.Equal(t, "province", result["ST"])
		assert.Equal(t, "DE", result["C"])
	})

	t.Run("should not panic on segment without equals sign", func(t *testing.T) {
		assert.NotPanics(t, func() {
			result := extractSubject("CN=agent,garbage")
			assert.Equal(t, "agent", result["CN"])
			assert.Empty(t, result["garbage"])
		})
	})

	t.Run("should not panic on single key without value", func(t *testing.T) {
		assert.NotPanics(t, func() {
			result := extractSubject("CN")
			assert.Empty(t, result["CN"])
		})
	})

	t.Run("should not panic on empty string", func(t *testing.T) {
		assert.NotPanics(t, func() {
			result := extractSubject("")
			assert.Empty(t, result)
		})
	})

	t.Run("should handle value containing equals sign", func(t *testing.T) {
		assert.NotPanics(t, func() {
			result := extractSubject("CN=test=value,O=Org")
			assert.Equal(t, "test=value", result["CN"])
			assert.Equal(t, "Org", result["O"])
		})
	})

	t.Run("should handle trailing comma", func(t *testing.T) {
		assert.NotPanics(t, func() {
			result := extractSubject("CN=agent,")
			assert.Equal(t, "agent", result["CN"])
		})
	})

	t.Run("should handle spaces around segments", func(t *testing.T) {
		result := extractSubject(" CN=test , O=Org ")
		assert.Equal(t, "test", result["CN"])
		assert.Equal(t, "Org", result["O"])
	})
}

func TestParseSubject(t *testing.T) {
	t.Run("should not panic on malformed subject", func(t *testing.T) {
		assert.NotPanics(t, func() {
			name := parseSubject("CN")
			assert.Empty(t, name.CommonName)
		})
	})

	t.Run("should not panic on subject with trailing comma", func(t *testing.T) {
		assert.NotPanics(t, func() {
			name := parseSubject("CN=runtime,")
			assert.Equal(t, "runtime", name.CommonName)
		})
	})

	t.Run("should parse valid subject into pkix.Name", func(t *testing.T) {
		name := parseSubject("O=Org,OU=OrgUnit,L=locality,ST=province,C=DE,CN=test")
		assert.Equal(t, "test", name.CommonName)
		assert.Equal(t, []string{"OrgUnit"}, name.OrganizationalUnit)
		assert.Equal(t, []string{"Org"}, name.Organization)
		assert.Equal(t, []string{"locality"}, name.Locality)
		assert.Equal(t, []string{"province"}, name.Province)
		assert.Equal(t, []string{"DE"}, name.Country)
	})
}
