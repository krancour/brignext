export enum LogLevel {
  ALL = 0,
  LOG,
  INFO,
  WARN,
  ERROR,
  NONE
}

export class Logger {
  context: string
  logLevel: LogLevel
  constructor(ctx: string[] | string = [], logLevel = LogLevel.LOG) {
    if (typeof ctx === "string") {
      ctx = [ctx]
    }
    this.context = `[${new Array("brigade", ...ctx).join(":")}]`
    this.logLevel = logLevel
  }
  error(message?: any, ...optionalParams: any[]): void {
    if (LogLevel.ERROR >= this.logLevel) {
      console.error(this.context, message, ...optionalParams)
    }
  }
  warn(message?: any, ...optionalParams: any[]): void {
    if (LogLevel.WARN >= this.logLevel) {
      console.warn(this.context, message, ...optionalParams)
    }
  }
  info(message?: any, ...optionalParams: any[]): void {
    if (LogLevel.INFO >= this.logLevel) {
      console.info(this.context, message, ...optionalParams)
    }
  }
  log(message?: any, ...optionalParams: any[]): void {
    if (LogLevel.LOG >= this.logLevel) {
      console.log(this.context, message, ...optionalParams)
    }
  }
}
