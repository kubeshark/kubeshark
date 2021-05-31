package sensitiveDataFiltering

const maskedFieldPlaceholderValue = "[REDACTED]"

//these values MUST be all lower case
var personallyIdentifiableDataFields = []string{"token", "authorization", "authentication", "cookie", "userid", "password",
	"username", "user", "key", "passcode", "pass", "auth", "authtoken", "jwt",
	"bearer", "clientid", "clientsecret", "redirecturi", "phonenumber",
	"zip", "zipcode", "address", "country", "firstname", "lastname",
	"middlename", "fname", "lname", "birthdate"}
