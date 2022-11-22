import Moment from 'moment';

const IP_ADDRESS_REGEX = /([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})(:([0-9]{1,5}))?/

type JSONValue =
  | string
  | number
  | boolean
  | Object


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

  static formatDate = (date) => {
    let d = new Date(date),
        month = '' + (d.getMonth() + 1),
        day = '' + d.getDate(),
        year = d.getFullYear();
    const hoursAndMinutes = Utils.getHoursAndMinutes(date);
    if (month.length < 2) 
        month = '0' + month;
    if (day.length < 2) 
        day = '0' + day;
    const newDate = [year, month, day].join('-');
    return [hoursAndMinutes, newDate].join(' ');
}

  static createUniqueObjArrayByProp = (objArray, prop) => {
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

  static downloadFile = (data: string, filename: string, fileType: string) => {
    const blob = new Blob([data], { type: fileType })
    const a = document.createElement('a');
    a.href = window.URL.createObjectURL(blob);
    a.download = filename;
    a.click();
    a.remove();
  }

  static exportToJson = (data: JSONValue, name) => {
    Utils.downloadFile(JSON.stringify(data), `${name}.json`, 'text/json')
  }

  static getTimeFormatted = (time: Moment.MomentInput) => {
    return Moment(time).utc().format('MM/DD/YYYY, h:mm:ss.SSS A')
  }

  static getNow = (format: string = 'MM/DD/YYYY, HH:mm:ss.SSS') => {
    return Moment().format(format)
  }
}
