package utils

const maskedFieldPlaceholderValue = "[REDACTED]"
var personallyIdentifiableDataFields = []string {"token", "authorization", "authentication", "cookie", "userid", "password",
												 "username", "user", "key", "passcode", "pass", "auth", "authtoken", "jwt",
												 "bearer", "clientid", "clientsecret", "redirecturi", "phonenumber",
												 "zip", "zipcode", "address", "country", "city", "state", "residence",
                                                 "name", "firstname", "lastname", "suffix", "middlename", "fname", "lname",
                                                 "mname", "birthday", "birthday", "birthdate", "bday", "sender", "receiver"}
