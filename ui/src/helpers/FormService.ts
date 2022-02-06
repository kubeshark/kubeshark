export class FormService{

    EMAIL_REGEXP = "^\\w+([-+.']\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"

    isValidEmail = (email : string) => {
        return new RegExp(this.EMAIL_REGEXP).test(email)
    }
}