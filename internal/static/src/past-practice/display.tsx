import dayjs from "dayjs";
import timezone from "dayjs/plugin/timezone";
import utc from "dayjs/plugin/utc";
import relativeTime from "dayjs/plugin/relativeTime";
import { useEffect, useState } from "preact/hooks";
import { Link } from "../ui/links";
import { cn } from "../common";

dayjs.extend(timezone);
dayjs.extend(utc);
dayjs.extend(relativeTime);

type SqlNullString = {
  String: string;
  Valid: boolean;
};

type ListRecentPracticeSessionsRow = {
  id: string;
  durationMinutes: number;
  date: number;
  practicePieceMeasures: SqlNullString;
  pieceTitle: SqlNullString;
  pieceId: SqlNullString;
  pieceComposer: SqlNullString;
  spotName: SqlNullString;
  spotId: SqlNullString;
  spotMeasures: SqlNullString;
  spotPieceId: SqlNullString;
  spotPieceTitle: SqlNullString;
};

type PracticePiece = {
  pieceId: string;
  pieceTitle: string;
  pieceMeasures: string;
};

type PracticeSpot = {
  spotId: string;
  spotName: string;
  spotMeasures: string;
  spotPieceId: string;
  spotPieceTitle: string;
};

type DisplaySession = {
  id: string;
  durationMinutes: number;
  date: dayjs.Dayjs;
  pieces: PracticePiece[];
  spots: PracticeSpot[];
};

export function PastPracticeDisplay({
  sessions,
  wide = false,
}: {
  sessions: string;
  wide?: boolean;
}) {
  const [displaySessions, setDisplaySessions] = useState<DisplaySession[]>([]);

  useEffect(
    function () {
      const rows = JSON.parse(sessions) as ListRecentPracticeSessionsRow[];
      const displaySessionMap = new Map<string, DisplaySession>();
      if (!(rows instanceof Array)) {
        return;
      }
      // track the seen ids so we don't double(or more) count the durations
      let seenIds = new Set();
      for (let row of rows) {
        const localDate = dayjs.unix(row.date).tz(dayjs.tz.guess());
        if (!displaySessionMap.get(localDate.format("YYYY-MM-DD"))) {
          displaySessionMap.set(localDate.format("YYYY-MM-DD"), {
            id: row.id,
            durationMinutes: 0,
            date: localDate,
            pieces: [],
            spots: [],
          });
        }
        // this is a mutable reference
        const displaySession = displaySessionMap.get(
          localDate.format("YYYY-MM-DD"),
        ) as DisplaySession;
        if (!seenIds.has(row.id)) {
          seenIds.add(row.id);
          displaySession.durationMinutes += row.durationMinutes;
        }
        if (
          row.pieceId.Valid &&
          !displaySession.pieces.find((p) => p.pieceId === row.pieceId.String)
        ) {
          displaySession.pieces.push({
            pieceId: row.pieceId.String,
            pieceTitle: row.pieceTitle.String,
            pieceMeasures: row.practicePieceMeasures.String,
          });
        }
        if (
          row.spotId.Valid &&
          !displaySession.spots.find((p) => p.spotId === row.spotId.String)
        ) {
          displaySession.spots.push({
            spotId: row.spotId.String,
            spotName: row.spotName.String,
            spotMeasures: row.spotMeasures.String,
            spotPieceId: row.spotPieceId.String,
            spotPieceTitle: row.spotPieceTitle.String,
          });
        }
      }
      const ds = [...displaySessionMap.values()].sort(
        (a, b) => b.date.unix() - a.date.unix(),
      );
      setDisplaySessions(ds.slice(0, 3));
    },
    [sessions],
  );
  if (displaySessions.length === 0) {
    return null;
  }
  return (
    <>
      <ul
        className={cn(
          "grid w-full list-none grid-cols-1 gap-2",
          wide && "md:grid-cols-2",
        )}
      >
        {displaySessions.map((ps) => (
          <li
            key={ps.id}
            className="flex flex-col rounded-xl bg-white/50 px-4 py-2 shadow shadow-neutral-700/5"
          >
            <h3 className="col-span-full text-lg font-semibold">
              Practiced for <HoursMinutesDisplay minutes={ps.durationMinutes} />{" "}
              - {ps.date.fromNow()}
            </h3>
            <div className="grid grid-cols-2 gap-x-2">
              <div>
                <h4 className="font-medium">Spots</h4>
                <ul>
                  {ps.spots.map((s) => {
                    return (
                      <li key={s.spotId}>
                        <Link
                          className="underline"
                          href={`/library/pieces/${s.spotPieceId}/spots/${s.spotId}`}
                        >
                          {s.spotName} - {s.spotPieceTitle}
                        </Link>
                      </li>
                    );
                  })}
                </ul>
              </div>
              <div>
                <h4 className="font-medium">Pieces</h4>
                <ul>
                  {ps.pieces.map((p) => {
                    return (
                      <li key={p.pieceId}>
                        <Link
                          href={`/library/pieces/${p.pieceId}`}
                          className="underline"
                        >
                          {p.pieceTitle} - {p.pieceMeasures}
                        </Link>
                      </li>
                    );
                  })}
                </ul>
              </div>
            </div>
          </li>
        ))}
      </ul>
    </>
  );
}

function HoursMinutesDisplay({ minutes }: { minutes: number }) {
  const hours = Math.floor(minutes / 60);
  const minutesRemainder = minutes % 60;
  if (hours === 0) {
    return <>{`${minutesRemainder} minutes`}</>;
  }
  if (minutesRemainder === 0) {
    return <>{`${hours} hours`}</>;
  }
  return <>{`${hours} hours and ${minutesRemainder} minutes`}</>;
}
