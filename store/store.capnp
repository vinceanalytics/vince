using Go = import "/go.capnp";
@0xda1eebf4d7c0f83e;
$Go.package("store");
$Go.import("store");


struct Calendar{
    visitors @0 :List(Float64);
    views @1 :List(Float64);
    events @2 :List(Float64);
    visits @3 :List(Float64);
    bounceRate @4 :List(Float64);
    visitDuration @5 :List(Float64);
    viewsPerVisit @6 :List(Float64);
}


