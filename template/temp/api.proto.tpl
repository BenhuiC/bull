syntax = "proto3";
package api;

option go_package="api/proto";

import "google/api/annotations.proto";
import "gogoproto/gogo.proto";
import "google/protobuf/wrappers.proto";

service {{ .ProjectName}}Service {
  // 探活
  rpc Ping(Empty) returns (Pong){
    option(google.api.http) ={
      get: "/ping"
    };
  }
}

message Empty {}

message Pong {
  Pong string =1;
}


