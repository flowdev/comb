// Comb project: comb & bug for logo

// Global resolution
$fs = 0.1; // Don't generate smaller facets than 0.1 mm
$fa = 5;   // Don't generate larger angles than 5 degrees

//   B U G :
bugDiameter = 10;
bugLegLenth = 3;
bugLegDiameter = 0.5;

bug();


module bug() {
  color("red") bugBody(bugDiameter);

  translate([0,0.5,0]) color("black") bugHead(bugDiameter);

  color("black") bugLeg1(bugDiameter, bugLegDiameter, bugLegLenth);
  color("black") bugLeg2(bugDiameter, bugLegDiameter, bugLegLenth);
  color("black") bugLeg3(bugDiameter, bugLegDiameter, bugLegLenth);

  color("black") mirror([1,0,0]) bugLeg1(bugDiameter, bugLegDiameter, bugLegLenth);
  color("black") mirror([1,0,0]) bugLeg2(bugDiameter, bugLegDiameter, bugLegLenth);
  color("black") mirror([1,0,0]) bugLeg3(bugDiameter, bugLegDiameter, bugLegLenth);
}

module bugLeg1(bugd, d, l) {
  translate([-bugd/1.65, bugd/1.5, 0])
    rotate([45,90,0]) {
      cylinder(h=l, d=d);
    }
  translate([-bugd/1.65, bugd/1.05, 0])
    rotate([90,90,0]) {
      cylinder(h=l, d=d);
    }
}
module bugLeg2(bugd, d, l) {
  translate([-bugd/1.5, 0, 0])
    rotate([0,90,0]) {
      cylinder(h=l, d=d);
    }
  translate([-bugd/1.26, bugd/3.6, 0])
    rotate([60,90,0]) {
      cylinder(h=l, d=d);
    }
}
module bugLeg3(bugd, d, l) {
  translate([-bugd/1.7, -bugd/1.5, 0])
    rotate([-45,90,0]) {
      cylinder(h=l, d=d);
    }
  translate([-bugd/1.25, -bugd/2.4, 0])
    rotate([45,90,0]) {
      cylinder(h=l, d=d);
    }
}

module bugBody(diam) {
  difference() {
    bugSphere(diam);
    translate([0,diam + diam/2,0])
      cube(diam*2,center=true);    
  }
}
module bugHead(diam) {
  difference() {
    bugSphere(diam);
    translate([0,-diam*1.55,0])
      cube(diam*4,center=true);
  }
}
module bugSphere(diam) {
  scale([1,1.5,0.5]) sphere(d=diam);
}