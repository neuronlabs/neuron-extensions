package migrate

func init() {
	panicer := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	panicer(registerDataType(FChar))
	panicer(registerDataType(FVarChar))
	panicer(registerDataType(FText))
	// Numerics
	panicer(registerDataType(FInteger))
	panicer(registerDataType(FSmallInt))
	panicer(registerDataType(FBigInt))
	panicer(registerDataType(FDecimal))
	panicer(registerDataType(FNumeric))
	panicer(registerDataType(FReal))
	panicer(registerDataType(FDouble))
	panicer(registerDataType(FSerial))
	panicer(registerDataType(FBigSerial))
	// Binary
	panicer(registerDataType(FBytea))
	panicer(registerDataType(FBoolean))
	// Date / Time
	panicer(registerDataType(FDate))
	panicer(registerDataType(FTimestamp))
	panicer(registerDataType(FTimestampTZ))
	panicer(registerDataType(FTime))
	panicer(registerDataType(FTimeTZ))

	// AddModel default TagSetterFunctions values - name, index
	TagSetterFunctions[TagDatabaseName] = NameSetter
	TagSetterFunctions[TagColumnIndex] = IndexSetter
	TagSetterFunctions[TagDataType] = DataTypeSetter
	TagSetterFunctions[cNotNull] = ConstraintSetter
	TagSetterFunctions[cUnique] = ConstraintSetter
	TagSetterFunctions[cForeign] = ConstraintSetter
}
