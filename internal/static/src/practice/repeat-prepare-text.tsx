import { ChevronRightIcon } from "@heroicons/react/24/solid";

export function RepeatPrepareText({ open = false }: { open?: boolean }) {
  return (
    <section id="repeat-prepare" className="w-full">
      <div className="flex w-full flex-col gap-2 pt-4 md:flex-row">
        <details
          open={open}
          className="prose prose-neutral flex h-min w-full flex-col rounded-xl bg-neutral-700/5 p-4 md:w-1/2"
        >
          <summary className="flex cursor-pointer items-center justify-between text-left text-2xl">
            Preparation
            <ChevronRightIcon className="summary-icon -mr-1 size-6 transition-transform" />
          </summary>
          <ul className="flex flex-grow flex-col justify-around text-lg">
            <li>
              Where <em className="italic">exactly</em> does your section start
              and stop?
            </li>
            <li>What makes it difficult?</li>
            <li>What will you think about before you play?</li>
            <li>What will you think about while you are playing?</li>
          </ul>
        </details>
        <details
          className="prose prose-neutral h-min w-full rounded-xl bg-neutral-700/5 p-4 md:w-1/2"
          open={open}
        >
          <summary className="flex cursor-pointer items-center justify-between text-left text-2xl">
            How it works
            <ChevronRightIcon className="summary-icon -mr-1 size-6 transition-transform" />
          </summary>
          <ul className="text-base">
            <li>The goal is to practice five times without a mistake.</li>
            <li>Practice as slowly as you need to be successful</li>
            <li>Take time between each repetition to reset.</li>
            <li>
              Avoid making things worse by taking a break if things aren’t going
              well.
            </li>
            <li>
              There’s a small timeout to keep you from going way too fast.
            </li>
            <li>When you’re ready, click the button below to get started.</li>
          </ul>
        </details>
      </div>
    </section>
  );
}
