import register from "preact-custom-element";
import { RandomSpots } from "./practice/random-spots";
import { SequenceSpots } from "./practice/sequence-spots";
import { Repeat } from "./practice/repeat";
import { StartingPoint } from "./practice/starting-point";

try {
  register(RandomSpots, "random-spots", [], { shadow: false });
  register(SequenceSpots, "sequence-spots", [], { shadow: false });
  register(Repeat, "repeat-practice", [], { shadow: false });
  register(StartingPoint, "starting-point", [], { shadow: false });
} catch (err) {
  console.log(err);
}
