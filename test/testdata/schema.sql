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
) PRIMARY KEY(PKey);

CREATE UNIQUE INDEX FullTypesByFTString ON FullTypes(FTString);

CREATE INDEX FullTypesByIntDate ON FullTypes(FTInt, FTDate);

CREATE INDEX FullTypesByIntTimestamp ON FullTypes(FTInt, FTTimestamp);

CREATE INDEX FullTypesByInTimestampNull ON FullTypes(FTInt, FTTimestampNull);

CREATE INDEX FullTypesByTimestamp ON FullTypes(FTTimestamp);

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
