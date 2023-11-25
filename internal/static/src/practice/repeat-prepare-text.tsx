export function RepeatPrepareText() {
  return (
    <>
      <div>
        <h1 className="py-1 text-left text-2xl font-bold">Repeat Practicing</h1>
        <p className="text-left text-base">
          Repeat practicing is an important part of learning, but you need to do
          it carefully!
        </p>
        <div className="grid w-full grid-cols-1 gap-2 py-4 md:grid-cols-2">
          <div className="prose prose-neutral flex h-full flex-col rounded-xl bg-neutral-700/5 p-4">
            <h3 className="text-left text-2xl">Answer these questions first</h3>
            <ul className="flex flex-grow flex-col justify-around text-lg">
              <li>
                Where <em className="italic">exactly</em> does your section
                start and stop?
              </li>
              <li>What makes it difficult?</li>
              <li>What will you think about prefore you play?</li>
              <li>What will you think about while you are playing?</li>
            </ul>
          </div>
          <div className="prose prose-neutral rounded-xl bg-neutral-700/5 p-4 md:grid-cols-2">
            <h3 className="text-left text-2xl">How it works</h3>
            <ul className="text-base">
              <li>The goal is to practice five times without a mistake.</li>
              <li>Practice as slowly as you need to be successful</li>
              <li>Take time between each repetition to reset.</li>
              <li>
                Avoid making things worse by taking a break if things aren’t
                going well.
              </li>
              <li>
                There’s a small timeout to keep you from going way too fast.
              </li>
              <li>When you’re ready, click the button below to get started.</li>
            </ul>
          </div>
        </div>
      </div>
    </>
  );
}
