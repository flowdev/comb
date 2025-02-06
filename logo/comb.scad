// Comb project: comb for logo

// Global resolution
$fs = 0.1; // Don't generate smaller facets than 0.1 mm
$fa = 5;   // Don't generate larger angles than 5 degrees

//   B U G :
combLen = 10;

comb();


module comb() {
  color("#a98") bar(combLen);
  for (i = [-combLen/2:combLen/5:combLen/2]) {
    color("#a98") translate([i,0,0]) tooth(combLen/5);
  }
}

module bar(len) {
  cube([len+1, len/7, 1], center=true);
}
module tooth(len) {
  translate([0,-len/4,0]) rotate([90,90,0]) cylinder(h=len, d=1);
  translate([0,-len*1.26,0]) sphere(d=1);
}