# gRPC-Gateway Support
protoc-gen-connect-openapi also has support for the [gRPC-Gateway annotations](https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/adding_annotations/) provided by the [google/api/annotations.proto](https://github.com/googleapis/googleapis/blob/master/google/api/annotations.proto). Here's an example of what this looks like in a protobuf file:

```protobuf
syntax = "proto3";

package io.swagger.petstore.v2;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";

service PetService {
  rpc GetPetByID(PetID) returns (Pet) {
    option (google.api.http).get = "/pet/{pet_id}";
    option idempotency_level = NO_SIDE_EFFECTS;
  }
  rpc UpdatePetWithForm(UpdatePetWithFormReq) returns (google.protobuf.Empty) {
    option (google.api.http).post = "/pet/{pet_id}";
    option (google.api.http).body = "*";
  }
  rpc DeletePet(PetID) returns (google.protobuf.Empty) {
    option (google.api.http).delete = "/pet/{pet_id}";
  }
  rpc UploadFile(UploadFileReq) returns (ApiResponse) {
    option (google.api.http).post = "/pet/{pet_id}/uploadImage";
    option (google.api.http).body = "*";
  }
  rpc AddPet(Pet) returns (Pet) {
    option (google.api.http).post = "/pet";
    option (google.api.http).body = "*";
  }
  rpc UpdatePet(Pet) returns (Pet) {
    option (google.api.http).put = "/pet";
    option (google.api.http).body = "*";
  }
  rpc FindPetsByTag(TagReq) returns (Pets) {
    option deprecated = true;
    option (google.api.http).get = "/pet/findByTags";
    option (google.api.http).response_body = "pets";
    option idempotency_level = NO_SIDE_EFFECTS;
  }
  rpc FindPetsByStatus(StatusReq) returns (Pets) {
    option (google.api.http).get = "/pet/findByStatus";
    option (google.api.http).response_body = "pets";
    option idempotency_level = NO_SIDE_EFFECTS;
  }
}
```

For more information on how to use each option in your Protobuf file, you can reference [the gRPC-Gateway documentation](https://github.com/grpc-ecosystem/grpc-gateway/blob/main/README.md) and the [Adding gRPC-Gateway annotations to an existing proto file](https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/adding_annotations/) article. Note that this is a new feature, so if find something that isn't supported that you need, please [create an issue](https://github.com/sudorandom/protoc-gen-connect-openapi/issues/new).
