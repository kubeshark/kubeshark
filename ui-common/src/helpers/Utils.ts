const IP_ADDRESS_REGEX = /([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})(:([0-9]{1,5}))?/


export class Utils {
  static isIpAddress = (address: string): boolean => IP_ADDRESS_REGEX.test(address)
  static lineNumbersInString = (code: string): number => code.split("\n").length;

  static humanFileSize(bytes, si = false, dp = 1) {
    const thresh = si ? 1000 : 1024;

    if (Math.abs(bytes) < thresh) {
      return bytes + ' B';
    }

    const units = si
      ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
      : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
    let u = -1;
    const r = 10 ** dp;

    do {
      bytes /= thresh;
      ++u;
    } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1);


    return bytes.toFixed(dp) + ' ' + units[u];
  }

  static padTo2Digits = (num) => {
    return String(num).padStart(2, '0');
  }

  static getHoursAndMinutes = (protocolTimeKey) => {
    const time = new Date(protocolTimeKey)
    const hoursAndMinutes = Utils.padTo2Digits(time.getHours()) + ':' + Utils.padTo2Digits(time.getMinutes());
    return hoursAndMinutes;
  }

  static creatUniqueObjArrayByProp = (objArray, prop) => {
    const map = new Map(objArray.map((item) => [item[prop], item])).values()
    return Array.from(map);
  }

  static isJson = (str) => {
    try {
      JSON.parse(str);
    } catch (e) {
      return false;
    }
    return true;
  }

  static stringToColor = (str) => {
    let colors = ["#e51c23", "#e91e63", "#9c27b0", "#673ab7", "#3f51b5", "#5677fc", "#03a9f4", "#00bcd4", "#009688", "#259b24", "#8bc34a", "#afb42b", "#ff9800", "#ff5722", "#795548", "#607d8b"]

    let hash = 0;
    if (str.length === 0) return hash;
    for (let i = 0; i < str.length; i++) {
      hash = str.charCodeAt(i) + ((hash << 5) - hash);
      hash = hash & hash;
    }
    hash = ((hash % colors.length) + colors.length) % colors.length;
    return colors[hash];
  }

}
