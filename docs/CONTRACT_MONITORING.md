# OpenAPI Specification (OAS) Contract Monitoring

An OAS/Swagger file can contain schemas under `parameters` and `responses` fields. With `--contract catalogue.yaml`
CLI option, you can pass your API description to Mizu and the traffic will automatically be validated
against the contracts.

Below is an example of an OAS/Swagger file from [Sock Shop](https://microservices-demo.github.io/) microservice demo
that contains a bunch contracts:

```yaml
openapi: 3.0.1
info:
  title: Catalogue resources
  version: 1.0.0
  description: ""
  license:
    name: MIT
    url: http://github.com/gruntjs/grunt/blob/master/LICENSE-MIT
paths:
  /catalogue:
    get:
      description: Catalogue API
      operationId: List catalogue
      responses:
        200:
          description: ""
          content:
            application/json;charset=UTF-8:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Listresponse'
  /catalogue/{id}:
    get:
      operationId: Get an item
      parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
        example: a0a4f044-b040-410d-8ead-4de0446aec7e
      responses:
        200:
          description: ""
          content:
            application/json; charset=UTF-8:
              schema:
                $ref: '#/components/schemas/Getanitemresponse'
  /catalogue/size:
    get:
      operationId: Get size
      responses:
        200:
          description: ""
          content:
            application/json;charset=UTF-8:
              schema:
                $ref: '#/components/schemas/Getsizeresponse'
  /tags:
    get:
      operationId: List_
      responses:
        200:
          description: ""
          content:
            application/json;charset=UTF-8:
              schema:
                $ref: '#/components/schemas/Listresponse3'
components:
  schemas:
    Listresponse:
      title: List response
      required:
      - count
      - description
      - id
      - imageUrl
      - name
      - price
      - tag
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        description:
          type: string
        imageUrl:
          type: array
          items:
            type: string
        price:
          type: number
          format: double
        count:
          type: integer
          format: int32
        tag:
          type: array
          items:
            type: string
    Getanitemresponse:
      title: Get an item response
      required:
      - count
      - description
      - id
      - imageUrl
      - name
      - price
      - tag
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        description:
          type: string
        imageUrl:
          type: array
          items:
            type: string
        price:
          type: number
          format: double
        count:
          type: integer
          format: int32
        tag:
          type: array
          items:
            type: string
    Getsizeresponse:
      title: Get size response
      required:
      - size
      type: object
      properties:
        size:
          type: integer
          format: int32
    Listresponse3:
      title: List response3
      required:
      - tags
      type: object
      properties:
        tags:
          type: array
          items:
            type: string
```

Pass it to Mizu through the CLI option: `mizu tap -n sock-shop --contract catalogue.yaml`

Now Mizu will monitor the traffic against these contracts.

If an entry fails to comply with the contract, it's marked with `Breach` notice in the UI.
The reason of the failure can be seen under the `CONTRACT` tab in the details layout.

### Notes

Make sure that you;

- specified the `openapi` version
- specified the `info.version` version in the YAML
- and removed `servers` field from the YAML

Otherwise the OAS file cannot be recognized. (see [this issue](https://github.com/getkin/kin-openapi/issues/356))
