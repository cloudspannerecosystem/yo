-- Copyright (c) 2020 Mercari, Inc.
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy of
-- this software and associated documentation files (the "Software"), to deal in
-- the Software without restriction, including without limitation the rights to
-- use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
-- the Software, and to permit persons to whom the Software is furnished to do so,
-- subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in all
-- copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
-- FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
-- COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
-- IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
-- CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

CREATE TABLE CompositePrimaryKeys (
  Id INT64 NOT NULL,
  PKey1 STRING(32) NOT NULL,
  PKey2 INT64 NOT NULL,
  Error INT64 NOT NULL,
  X STRING(32) NOT NULL,
  Y STRING(32) NOT NULL,
  Z STRING(32) NOT NULL,
) PRIMARY KEY(PKey1, PKey2);

CREATE INDEX CompositePrimaryKeysByXY     ON CompositePrimaryKeys(X, Y);
CREATE INDEX CompositePrimaryKeysByError  ON CompositePrimaryKeys(Error);
CREATE INDEX CompositePrimaryKeysByError2 ON CompositePrimaryKeys(Error) STORING(Z);
CREATE INDEX CompositePrimaryKeysByError3 ON CompositePrimaryKeys(Error) STORING(Z, Y);

CREATE TABLE CustomCompositePrimaryKeys (
  Id INT64 NOT NULL,
  PKey1 STRING(32) NOT NULL,
  PKey2 INT64 NOT NULL,
  Error INT64 NOT NULL,
  X STRING(32) NOT NULL,
  Y STRING(32) NOT NULL,
  Z STRING(32) NOT NULL,
) PRIMARY KEY(PKey1, PKey2);

CREATE INDEX CustomCompositePrimaryKeysByXY     ON CustomCompositePrimaryKeys(X, Y);
CREATE INDEX CustomCompositePrimaryKeysByError  ON CustomCompositePrimaryKeys(Error);
CREATE INDEX CustomCompositePrimaryKeysByError2 ON CustomCompositePrimaryKeys(Error) STORING(Z);
CREATE INDEX CustomCompositePrimaryKeysByError3 ON CustomCompositePrimaryKeys(Error) STORING(Z, Y);

CREATE TABLE OutOfOrderPrimaryKeys (
  PKey1 STRING(32) NOT NULL,
  PKey2 STRING(32) NOT NULL,
  PKey3 STRING(32) NOT NULL,
) PRIMARY KEY(PKey2, PKey1, PKey3);

CREATE TABLE FullTypes (
  PKey STRING(32) NOT NULL,
  FTString STRING(32) NOT NULL,
  FTStringNull STRING(32),
  FTBool BOOL NOT NULL,
  FTBoolNull BOOL,
  FTBytes BYTES(32) NOT NULL,
  FTBytesNull BYTES(32),
  FTTimestamp TIMESTAMP NOT NULL,
  FTTimestampNull TIMESTAMP,
  FTInt INT64 NOT NULL,
  FTIntNull INT64,
  FTFloat FLOAT64 NOT NULL,
  FTFloatNull FLOAT64,
  FTDate DATE NOT NULL,
  FTDateNull DATE,
  FTJson JSON NOT NULL,
  FTJsonNull JSON,
  FTArrayStringNull ARRAY<STRING(32)>,
  FTArrayString ARRAY<STRING(32)> NOT NULL,
  FTArrayBoolNull ARRAY<BOOL>,
  FTArrayBool ARRAY<BOOL> NOT NULL,
  FTArrayBytesNull ARRAY<BYTES(32)>,
  FTArrayBytes ARRAY<BYTES(32)> NOT NULL,
  FTArrayTimestampNull ARRAY<TIMESTAMP>,
  FTArrayTimestamp ARRAY<TIMESTAMP> NOT NULL,
  FTArrayIntNull ARRAY<INT64>,
  FTArrayInt ARRAY<INT64> NOT NULL,
  FTArrayFloatNull ARRAY<FLOAT64>,
  FTArrayFloat ARRAY<FLOAT64> NOT NULL,
  FTArrayDateNull ARRAY<DATE>,
  FTArrayDate ARRAY<DATE> NOT NULL,
  FTArrayJsonNull ARRAY<JSON>,
  FTArrayJson ARRAY<JSON> NOT NULL,
) PRIMARY KEY(PKey);

CREATE UNIQUE INDEX FullTypesByFTString ON FullTypes(FTString);

CREATE INDEX FullTypesByIntDate ON FullTypes(FTInt, FTDate);

CREATE INDEX FullTypesByIntTimestamp ON FullTypes(FTInt, FTTimestamp);

CREATE INDEX FullTypesByInTimestampNull ON FullTypes(FTInt, FTTimestampNull);

CREATE INDEX FullTypesByTimestamp ON FullTypes(FTTimestamp);

CREATE TABLE CustomPrimitiveTypes (
  PKey STRING(32) NOT NULL,
  FTInt64 INT64 NOT NULL,
  FTInt64Null INT64,
  FTInt32 INT64 NOT NULL,
  FTInt32Null INT64,
  FTInt16 INT64 NOT NULL,
  FTInt16Null INT64,
  FTInt8 INT64 NOT NULL,
  FTInt8Null INT64,
  FTUInt64 INT64 NOT NULL,
  FTUInt64Null INT64,
  FTUInt32 INT64 NOT NULL,
  FTUInt32Null INT64,
  FTUInt16 INT64 NOT NULL,
  FTUInt16Null INT64,
  FTUInt8 INT64 NOT NULL,
  FTUInt8Null INT64,
  FTArrayInt64 ARRAY<INT64> NOT NULL,
  FTArrayInt64Null ARRAY<INT64>,
  FTArrayInt32 ARRAY<INT64> NOT NULL,
  FTArrayInt32Null ARRAY<INT64>,
  FTArrayInt16 ARRAY<INT64> NOT NULL,
  FTArrayInt16Null ARRAY<INT64>,
  FTArrayInt8 ARRAY<INT64> NOT NULL,
  FTArrayInt8Null ARRAY<INT64>,
  FTArrayUInt64 ARRAY<INT64> NOT NULL,
  FTArrayUInt64Null ARRAY<INT64>,
  FTArrayUInt32 ARRAY<INT64> NOT NULL,
  FTArrayUInt32Null ARRAY<INT64>,
  FTArrayUInt16 ARRAY<INT64> NOT NULL,
  FTArrayUInt16Null ARRAY<INT64>,
  FTArrayUInt8 ARRAY<INT64> NOT NULL,
  FTArrayUInt8Null ARRAY<INT64>,
) PRIMARY KEY(PKey);

CREATE TABLE MaxLengths (
  MaxString STRING(MAX) NOT NULL,
  MaxBytes BYTES(MAX) NOT NULL,
) PRIMARY KEY(MaxString);

CREATE TABLE snake_cases (
  id INT64 NOT NULL,
  string_id STRING(32) NOT NULL,
  foo_bar_baz INT64 NOT NULL,
) PRIMARY KEY(id);

CREATE INDEX snake_cases_by_string_id ON snake_cases(string_id, foo_bar_baz);

CREATE TABLE Items (
  ID INT64 NOT NULL,
  Price INT64 NOT NULL,
) PRIMARY KEY (ID);

CREATE TABLE FereignItems (
  ID INT64 NOT NULL,
  ItemID INT64 NOT NULL,
  Category INT64 NOT NULL,
  CONSTRAINT FK_ItemID_ForeignItems FOREIGN KEY (ItemID) REFERENCES Items (ID)
) PRIMARY KEY (ID);

CREATE TABLE GeneratedColumns (
  ID INT64 NOT NULL,
  FirstName STRING(50) NOT NULL,
  LastName STRING(50) NOT NULL,
  FullName STRING(100) NOT NULL AS (ARRAY_TO_STRING([FirstName, LastName], " ")) STORED,
) PRIMARY KEY (ID);

CREATE TABLE Inflectionzz (
  X STRING(32) NOT NULL,
  Y STRING(32) NOT NULL,
) PRIMARY KEY(X);

CREATE TABLE FullTextSearch (
  ID INT64 NOT NULL,
  Content STRING(2048) NOT NULL,
  Content_Tokens TOKENLIST AS (TOKENIZE_FULLTEXT(Content)) HIDDEN,
) PRIMARY KEY(ID);
