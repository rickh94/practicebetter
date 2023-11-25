import register from "preact-custom-element";
import { Chart } from "./about/chart";

try {
  register(Chart, "about-chart", [], { shadow: false });
} catch (err) {
  console.log(err);
}
