using Go = import "/go.capnp";
@0xda1eebf4d7c0f83e;
$Go.package("store");
$Go.import("store");


struct Calendar{
    visitors @0 :List(Float64);
    visits @1 :List(Float64);
    views @2 :List(Float64);
    events @3 :List(Float64);
}

struct Sum{
    # events metrics 
    visitors @0 :Float64;
    views @1 :Float64;
    events @2 :Float64; 

    # session metrics 
    visits @3 :Float64;
    bounceRate @4 :Float64;
    visitDuration @5 :Float64;
    viewsPerVisit @6 :Float64;
}

