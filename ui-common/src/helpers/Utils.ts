const IP_ADDRESS_REGEX = /([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})(:([0-9]{1,5}))?/


export class Utils {
    static isIpAddress = (address: string): boolean => IP_ADDRESS_REGEX.test(address)
    static lineNumbersInString = (code:string): number => code.split("\n").length;
}
