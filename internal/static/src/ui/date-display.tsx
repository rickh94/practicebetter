import dayjs from "dayjs";
import timezone from "dayjs/plugin/timezone";
import utc from "dayjs/plugin/utc";
import relativeTime from "dayjs/plugin/relativeTime";

dayjs.extend(timezone);
dayjs.extend(utc);
dayjs.extend(relativeTime);

export function DateFromNow({ epoch }: { epoch: string }) {
  const dt = dayjs.unix(parseInt(epoch, 10)).tz(dayjs.tz.guess());
  return <>{dt.fromNow()}</>;
}

export function NumberDate({ epoch }: { epoch: string }) {
  const dt = dayjs.unix(parseInt(epoch, 10)).tz(dayjs.tz.guess());
  return <time dateTime={dt.toISOString()}>{dt.format("YYYY-MM-DD")}</time>;
}

export function PrettyDate({ epoch }: { epoch: string }) {
  const dt = dayjs.unix(parseInt(epoch, 10)).tz(dayjs.tz.guess());
  return (
    <time dateTime={dt.toISOString()}>{dt.format("dddd MMM D, YYYY")}</time>
  );
}
