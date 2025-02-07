// Comb project: comb for logo

// Global resolution
$fs = 0.01; // Don't generate smaller facets than 0.01 mm
$fa = 1;   // Don't generate larger angles than 1 degrees

combLen = 20;
combCol = "#cc8877";

comb();

module comb() {
  color(combCol) bar(combLen);
  for (i = [-combLen/2:combLen/32:combLen/2]) {
    color(combCol) translate([i,0,0]) tooth(combLen/5);
  }
}

module bar(len) {
  cube([len+0.4, len/7, 0.3], center=true);
}
module tooth(len) {
  translate([0,-len/4,0])
    rotate([90,90,0])
    scale([0.1,0.4,1])
    cylinder(h=len, d=1);
  translate([0,-len*1.26,0])
    scale([0.4,1,0.1])
    sphere(d=1);
}