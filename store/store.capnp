using Go = import "/go.capnp";
@0xda1eebf4d7c0f83e;
$Go.package("store");
$Go.import("store");


struct Calendar{
    timestamps @0 :List(Int64);
    visitors @1 :List(Float64);
    views @2 :List(Float64);
    events @3 :List(Float64);
    visits @4 :List(Float64);
    bounceRate @5 :List(Float64);
    visitDuration @6 :List(Float64);
    viewsPerVisit @7 :List(Float64);
}


