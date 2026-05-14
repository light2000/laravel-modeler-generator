package meta

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/light2000/laravel-modeler-generator/common"
	"github.com/light2000/laravel-modeler-generator/proto"
)

var (
	VarCharDefaultLen      int32 = 255
	VarCharIndexDefaultLen int32 = 191
	VarCharMaxLen          int32 = 16383
	EnumDefaultLen         int32 = 32
	FileDefaultLen         int32 = 1024
)

type Attribute struct {
	*proto.Attribute
	// Add your custom fields below
	Project *Project
	Module  *Module
	Item    *Item
}

func FromProtoAttribute(p *proto.Attribute, item *Item, module *Module, project *Project) *Attribute {
	if p == nil {
		return nil
	}
	attr := &Attribute{
		Attribute: p,
		Project:   project,
		Module:    module,
		Item:      item,
	}

	project.MapAttributes[p.Id] = attr

	return attr
}

func (attr *Attribute) Dict() *Dict {
	if attr.DictId != "" {
		if _, ok := attr.Project.MapDict[attr.DictId]; !ok {
			panic(fmt.Sprintf("DictId %s not found in Project.MapDict", attr.DictId))
		}
		return attr.Project.MapDict[attr.DictId]
	}
	return nil
}

func (attr *Attribute) Snake() string {
	return attr.Code
}

func (attr *Attribute) Var() string {
	return common.ToLowerCamel(attr.Snake())
}

func (attr *Attribute) Class() string {
	return common.ToCamel(attr.Snake())
}

func (attr *Attribute) IsInteger() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_INT == attr.Type
}

func (attr *Attribute) IsDecimal() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_DECIMAL == attr.Type
}

func (attr *Attribute) IsBool() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_BOOL == attr.Type
}

func (attr *Attribute) IsString() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_STRING == attr.Type
}

func (attr *Attribute) IsText() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_TEXT == attr.Type
}

func (attr *Attribute) IsLongText() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_LONG_TEXT == attr.Type
}

func (attr *Attribute) IsDate() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_DATE == attr.Type
}

func (attr *Attribute) IsTime() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_TIME == attr.Type
}

func (attr *Attribute) IsDateTime() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_DATETIME == attr.Type
}

func (attr *Attribute) IsYear() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_YEAR == attr.Type
}

func (attr *Attribute) IsEnumOrSets() bool {
	isEnumOrSets := proto.AttributeType_ATTRIBUTE_TYPE_ENUM == attr.Type || proto.AttributeType_ATTRIBUTE_TYPE_SETS == attr.Type
	if isEnumOrSets && attr.DictId == "" {
		panic(fmt.Errorf("attribute %s.%s.%s is enum but dict_id is empty", attr.Module.Snake(), attr.Item.Snake(), attr.Snake()))
	}
	return isEnumOrSets
}

func (attr *Attribute) IsEnum() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_ENUM == attr.Type
}

func (attr *Attribute) IsSets() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_SETS == attr.Type
}

func (attr *Attribute) IsFile() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_FILE == attr.Type
}

func (attr *Attribute) IsFiles() bool {
	return proto.AttributeType_ATTRIBUTE_TYPE_FILES == attr.Type
}

func (attr *Attribute) IsFk() bool {
	return proto.AttributeBehavior_ATTRIBUTE_BEHAVIOR_FK == attr.Behavior
}

func (attr *Attribute) IsMorphId() bool {
	return proto.AttributeBehavior_ATTRIBUTE_BEHAVIOR_MORPH_ID == attr.Behavior
}

func (attr *Attribute) IsMorphType() bool {
	return proto.AttributeBehavior_ATTRIBUTE_BEHAVIOR_MORPH_TYPE == attr.Behavior
}

func (attr *Attribute) IsPassword() bool {
	return attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_PASSWORD)
}

func (attr *Attribute) IsHidden() bool {
	return attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_HIDDEN)
}

func (attr *Attribute) IsRelationField() bool {
	return proto.AttributeBehavior_ATTRIBUTE_BEHAVIOR_FK == attr.Behavior ||
		proto.AttributeBehavior_ATTRIBUTE_BEHAVIOR_MORPH_ID == attr.Behavior ||
		proto.AttributeBehavior_ATTRIBUTE_BEHAVIOR_MORPH_TYPE == attr.Behavior
}

func (attr *Attribute) HasStaticDefValue() bool {
	return attr.IsSets() || attr.IsFiles()
}

func (attr *Attribute) HasAbility(ability proto.AttributeAbility) bool {
	return slices.Contains(attr.Abilities, ability)
}

func (attr *Attribute) StaticDefValue() string {
	if attr.IsSets() {
		if proto.AttributeDefaultValueType_ATTRIBUTE_DEFAULT_VALUE_TYPE_VALUE == attr.DefaultValueType {
			optIds := strings.Split(attr.DefaultValue, ",")
			optCodes := make([]string, 0)
			for _, optId := range optIds {
				for _, opt := range attr.Dict().Options {
					if optId == opt.Id {
						optCodes = append(optCodes, opt.Code)
					}
				}
			}
			if len(optCodes) > 0 {
				return fmt.Sprintf("'['%s']'", strings.Join(optCodes, "', '"))
			}
		}
		return "'[]'"
	}

	if attr.IsFiles() {
		return "'[]'"
	}

	return ""
}

func (attr *Attribute) MorphItems() []*Item {
	if attr.Behavior != proto.AttributeBehavior_ATTRIBUTE_BEHAVIOR_MORPH_ID {
		panic(fmt.Errorf("attribute %s.%s.%s is not a morph id", attr.Module.Snake(), attr.Item.Snake(), attr.Snake()))
	}

	items := make([]*Item, 0)
	for _, relation := range attr.Item.Relations {
		if relation.AttrId == attr.Id && len(relation.MorphTargets) > 0 {
			for _, morphTarget := range relation.MorphTargets {
				items = append(items, attr.Project.GetItemById(morphTarget.TargetItemId))
			}
		}
	}
	return items
}

func (attr *Attribute) MorphEnumArr() []string {
	enumFields := make([]string, 0)
	for _, item := range attr.MorphItems() {
		enumFields = append(enumFields, item.MorphMapKey())
	}

	return enumFields
}

func (attr *Attribute) IsInIndex(isSingle bool) bool {
	for _, idx := range attr.Item.Indexes {
		if isSingle {
			if len(idx.AttrIds) == 1 && idx.AttrIds[0] == attr.Id {
				return true
			}
		} else if slices.Contains(idx.AttrIds, attr.Id) {
			return true
		}
	}
	return false
}

func (attr *Attribute) IsInUniqueIndex(isSingle bool) bool {
	for _, idx := range attr.Item.Indexes {
		if !idx.IsUnique() {
			continue
		}
		if isSingle {
			if len(idx.AttrIds) == 1 && idx.AttrIds[0] == attr.Id {
				return true
			}
		} else if slices.Contains(idx.AttrIds, attr.Id) {
			return true
		}
	}
	return false
}

func (attr *Attribute) IsSingleUnique() bool {
	return attr.IsInUniqueIndex(true)
}

func (attr *Attribute) MorphAble() string {
	if attr.Behavior != proto.AttributeBehavior_ATTRIBUTE_BEHAVIOR_MORPH_ID {
		panic(fmt.Errorf("attribute %s.%s.%s is not a morph id type", attr.Module.Snake(), attr.Item.Snake(), attr.Snake()))
	}
	return strings.TrimSuffix(attr.Snake(), "_id")
}

func (attr *Attribute) PhpMigration() string {
	field := fmt.Sprintf("'%s'", attr.Snake())
	comment := common.PhpEscape(attr.Name)
	migration := ""

	nullable := ""
	if attr.Nullable {
		nullable = "->nullable()"
	}

	if attr.DictId != "" && len(attr.Dict().Options) == 0 {
		nullable = "->nullable()"
	}

	if attr.IsSets() || attr.IsFiles() {
		return fmt.Sprintf("$table->json(%s)%s->comment(\"%s\");", field, nullable, comment)
	}

	defaultValue := ""
	if attr.DefaultValueType == proto.AttributeDefaultValueType_ATTRIBUTE_DEFAULT_VALUE_TYPE_VALUE {
		if proto.AttributeType_ATTRIBUTE_TYPE_INT == attr.Type || proto.AttributeType_ATTRIBUTE_TYPE_DECIMAL == attr.Type || proto.AttributeType_ATTRIBUTE_TYPE_BOOL == attr.Type {
			defaultValue = fmt.Sprintf("->default(%s)", attr.DefaultValue)
		} else if attr.IsEnum() {
			for _, opt := range attr.Dict().Options {
				if attr.DefaultValue == opt.Id {
					defaultValue = fmt.Sprintf("->default('%s')", opt.Code)
					break
				}
			}
		} else {
			defaultValue = fmt.Sprintf("->default('%s')", attr.DefaultValue)
		}
	} else if attr.DefaultValueType == proto.AttributeDefaultValueType_ATTRIBUTE_DEFAULT_VALUE_TYPE_CURRENT_TIMESTAMP && proto.AttributeType_ATTRIBUTE_TYPE_DATETIME == attr.Type {
		defaultValue = "->useCurrent()"
	} else if attr.DefaultValueType == proto.AttributeDefaultValueType_ATTRIBUTE_DEFAULT_VALUE_TYPE_NULL {
		nullable = "->nullable()"
		defaultValue = "->default(null)"
	}

	switch attr.Type {
	case proto.AttributeType_ATTRIBUTE_TYPE_INT:
		if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_INT == attr.FieldType {
			migration = fmt.Sprintf("$table->integer(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_TINYINT == attr.FieldType {
			migration = fmt.Sprintf("$table->tinyInteger(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_SMALLINT == attr.FieldType {
			migration = fmt.Sprintf("$table->smallInteger(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_MEDIUMINT == attr.FieldType {
			migration = fmt.Sprintf("$table->mediumInteger(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_BIGINT == attr.FieldType {
			migration = fmt.Sprintf("$table->bigInteger(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
		} else {
			migration = fmt.Sprintf("$table->integer(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
		}
	case proto.AttributeType_ATTRIBUTE_TYPE_DECIMAL:
		if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_DECIMAL == attr.FieldType {
			migration = fmt.Sprintf("$table->decimal(%s, %d, %d)%s%s->comment(\"%s\");", field, attr.Precision, attr.Scale, nullable, defaultValue, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_FLOAT == attr.FieldType {
			migration = fmt.Sprintf("$table->float(%s, %d, %d)%s%s->comment(\"%s\");", field, attr.Precision, attr.Scale, nullable, defaultValue, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_DOUBLE == attr.FieldType {
			migration = fmt.Sprintf("$table->double(%s, %d, %d)%s%s->comment(\"%s\");", field, attr.Precision, attr.Scale, nullable, defaultValue, comment)
		} else {
			migration = fmt.Sprintf("$table->decimal(%s, %d, %d)%s%s->comment(\"%s\");", field, attr.Precision, attr.Scale, nullable, defaultValue, comment)
		}
	case proto.AttributeType_ATTRIBUTE_TYPE_STRING, proto.AttributeType_ATTRIBUTE_TYPE_TEXT, proto.AttributeType_ATTRIBUTE_TYPE_LONG_TEXT:
		if attr.FieldType == proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_UNSPECIFIED {
			if attr.Type == proto.AttributeType_ATTRIBUTE_TYPE_STRING {
				attr.FieldType = proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_VARCHAR
			} else if attr.Type == proto.AttributeType_ATTRIBUTE_TYPE_TEXT {
				attr.FieldType = proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_TEXT
			} else if attr.Type == proto.AttributeType_ATTRIBUTE_TYPE_LONG_TEXT {
				attr.FieldType = proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_MEDIUMTEXT
			}
		}

		if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_CHAR == attr.FieldType {
			migration = fmt.Sprintf("$table->char(%s, %d)%s%s->comment(\"%s\");", field, attr.StrFieldLength(), nullable, defaultValue, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_TEXT == attr.FieldType {
			migration = fmt.Sprintf("$table->text(%s)%s->comment(\"%s\");", field, nullable, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_VARCHAR == attr.FieldType {
			if VarCharDefaultLen == attr.StrFieldLength() {
				migration = fmt.Sprintf("$table->string(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
			} else {
				migration = fmt.Sprintf("$table->string(%s, %d)%s%s->comment(\"%s\");", field, attr.StrFieldLength(), nullable, defaultValue, comment)
			}
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_MEDIUMTEXT == attr.FieldType {
			migration = fmt.Sprintf("$table->mediumText(%s)%s->comment(\"%s\");", field, nullable, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_TINYTEXT == attr.FieldType {
			migration = fmt.Sprintf("$table->tinyText(%s)%s->comment(\"%s\");", field, nullable, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_LONGTEXT == attr.FieldType {
			migration = fmt.Sprintf("$table->longText(%s)%s->comment(\"%s\");", field, nullable, comment)
		} else {
			panic(fmt.Errorf("no way migration use %d:%d attr type", attr.Type, attr.FieldType))
		}
	case proto.AttributeType_ATTRIBUTE_TYPE_BOOL:
		migration = fmt.Sprintf("$table->boolean(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
	case proto.AttributeType_ATTRIBUTE_TYPE_DATETIME:
		useCurrentOnUpdate := ""
		if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_DATETIME == attr.FieldType {
			migration = fmt.Sprintf("$table->dateTime(%s)%s%s%s->comment(\"%s\");", field, useCurrentOnUpdate, defaultValue, nullable, comment)
		} else if proto.AttributeFieldType_ATTRIBUTE_FIELD_TYPE_TIMESTAMP == attr.FieldType {
			migration = fmt.Sprintf("$table->timestamp(%s)%s%s%s->comment(\"%s\");", field, useCurrentOnUpdate, defaultValue, nullable, comment)
		} else {
			migration = fmt.Sprintf("$table->dateTime(%s)%s%s%s->comment(\"%s\");", field, useCurrentOnUpdate, defaultValue, nullable, comment)
		}
	case proto.AttributeType_ATTRIBUTE_TYPE_DATE:
		migration = fmt.Sprintf("$table->date(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
	case proto.AttributeType_ATTRIBUTE_TYPE_TIME:
		migration = fmt.Sprintf("$table->time(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
	case proto.AttributeType_ATTRIBUTE_TYPE_YEAR:
		migration = fmt.Sprintf("$table->year(%s)%s%s->comment(\"%s\");", field, nullable, defaultValue, comment)
	case proto.AttributeType_ATTRIBUTE_TYPE_ENUM:
		comments := make([]string, 0)
		for _, option := range attr.Dict().Options {
			comments = append(comments, fmt.Sprintf("%s(%s)", common.PhpEscape(option.Name), common.PhpEscape(option.Code)))
		}
		migration = fmt.Sprintf("$table->string(%s, %d)%s%s->comment(\"%s\");", field, attr.StrFieldLength(), nullable, defaultValue, comment+":"+strings.Join(comments, ","))
	case proto.AttributeType_ATTRIBUTE_TYPE_FILE:
		migration = fmt.Sprintf("$table->string(%s, %d)%s->comment(\"%s\");", field, attr.StrFieldLength(), nullable, comment)
	}

	if migration == "" {
		panic(fmt.Errorf("no way migration use %d attr type", attr.Type))
	}

	return migration
}

func (attr *Attribute) IsSequenceFakeField() bool {
	return attr.SequencePhpFakeValue() != ""
}

func (attr *Attribute) SequencePhpFakeValue() string {
	if attr.IsInUniqueIndex(false) {
		if attr.IsEnum() {
			return fmt.Sprintf("%s[$sequence->index]", attr.Dict().OptionPhpArray())
		}

		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_ACCOUNT) {
			return "'user_' .(1000 + $sequence->index)"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_NICKNAME) {
			return "fake()->name() . '_' . $sequence->index"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_AVATAR) {
			return "fake()->imageUrl(60, 60, 'cats') . '?random=' . $sequence->index"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_EMAIL) {
			return "'user_'. $sequence->index .'@example.com'"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_PHONE) {
			return "13800138000 + $sequence->index"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_URL) {
			return "fake()->url() . '?random=' . $sequence->index"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_IP) {
			return "'127.0.0.'. $sequence->index"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_COLOR) {
			return "'#'. str_pad(dechex($sequence->index), 6, '0', STR_PAD_LEFT)"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_IMAGE) {
			return "'https://example.com/'. $sequence->index . '.jpg'"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_VIDEO) {
			return "'https://example.com/'. $sequence->index . '.mp4'"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_AUDIO) {
			return "'https://example.com/'. $sequence->index . '.mp3'"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_DOCUMENT) {
			return "'https://example.com/'. $sequence->index . '.pdf'"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_ARCHIVE) {
			return "'https://example.com/'. $sequence->index . '.zip'"
		}

		if attr.Type == proto.AttributeType_ATTRIBUTE_TYPE_INT {
			return "$sequence->index"
		}
		if attr.Type == proto.AttributeType_ATTRIBUTE_TYPE_DECIMAL {
			factor := int(math.Pow10(int(attr.Scale)))
			return fmt.Sprintf(
				"sprintf('%%.%df', $sequence->index / %d)",
				attr.Scale,
				factor,
			)
		}
		if proto.AttributeType_ATTRIBUTE_TYPE_BOOL == attr.Type {
			return "$sequence->index % 2 == 0"
		}

		if proto.AttributeType_ATTRIBUTE_TYPE_STRING == attr.Type {
			return "fake()->lexify(str_repeat('?', rand(5, 10))) . '_' . $sequence->index"
		}
		if proto.AttributeType_ATTRIBUTE_TYPE_TEXT == attr.Type {
			return "fake()->text(rand(5, 10)) . '_' . $sequence->index"
		}
		if proto.AttributeType_ATTRIBUTE_TYPE_LONG_TEXT == attr.Type {
			return "fake()->text(rand(10, 20)) . '_' . $sequence->index"
		}

		if proto.AttributeType_ATTRIBUTE_TYPE_DATETIME == attr.Type {
			return "date('Y-m-d H:i:s', strtotime('+'.$sequence->index.' seconds'))"
		}

		if proto.AttributeType_ATTRIBUTE_TYPE_DATE == attr.Type {
			return "date('Y-m-d', strtotime('+'.$sequence->index.' days'))"
		}

		if proto.AttributeType_ATTRIBUTE_TYPE_TIME == attr.Type {
			return "date('H:i:s', strtotime('+'.$sequence->index.' seconds'))"
		}

		if proto.AttributeType_ATTRIBUTE_TYPE_YEAR == attr.Type {
			return "date('Y', strtotime('+'.$sequence->index.' years'))"
		}
		if proto.AttributeType_ATTRIBUTE_TYPE_FILE == attr.Type {
			return "'https://example.com/'. $sequence->index . '.zip'"
		}

		panic(fmt.Errorf("no way sequence fake value use %d attr type", attr.Type))
	}

	if attr.IsFk() {
		if attr.Nullable {
			return "null"
		}
		return "$sequence->index"
	}
	if attr.IsMorphId() {
		if attr.Nullable {
			return "null"
		}
		return "$sequence->index"
	}

	return ""
}

func (attr *Attribute) MorphTypeFakeValue() string {
	if attr.Nullable {
		return "null"
	}
	return fmt.Sprintf("['%s'][$sequence->index %% %d]", strings.Join(attr.MorphEnumArr(), "', '"), len(attr.MorphEnumArr()))
}

func (attr *Attribute) MorphType() string {
	return attr.MorphAble() + "_type"
}

func (attr *Attribute) MorphName() string {
	return strings.ReplaceAll(attr.Name, "ID", "") + "类型"
}

func (attr *Attribute) IsFakeField() bool {
	if attr.IsRelationField() {
		return false
	}

	if attr.IsSequenceFakeField() {
		return false
	}

	return true
}

func parseScaledDecimal(s string, scale int) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty")
	}

	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}

	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	if intPart == "" {
		intPart = "0"
	}

	if len(fracPart) > scale {
		fracPart = fracPart[:scale]
	} else {
		fracPart += strings.Repeat("0", scale-len(fracPart))
	}

	n, err := strconv.Atoi(intPart + fracPart)
	if err != nil {
		return 0, err
	}
	if neg {
		n = -n
	}
	return n, nil
}

func (attr *Attribute) PhpFakeValue() string {
	fakeVar := ""
	unique := ""
	switch attr.Type {
	case proto.AttributeType_ATTRIBUTE_TYPE_INT:
		min, err := strconv.Atoi(attr.FakeMin)
		if nil != err {
			min = 0
		}
		max, err := strconv.Atoi(attr.FakeMax)
		if nil != err {
			max = 999
		}
		fakeVar = fmt.Sprintf("fake()%s->numberBetween(%d, %d)", unique, min, max)
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_PHONE) {
			fakeVar = fmt.Sprintf("fake()%s->numberBetween(13000000000, 19999999999)", unique)
		}
	case proto.AttributeType_ATTRIBUTE_TYPE_DECIMAL:
		factor := int(math.Pow10(int(attr.Scale)))
		minInteger := 0
		maxInteger := int(math.Pow10(int(attr.Precision))) - 1

		if v, err := parseScaledDecimal(attr.FakeMin, int(attr.Scale)); err == nil {
			if v > minInteger {
				minInteger = v
			}
		}
		if v, err := parseScaledDecimal(attr.FakeMax, int(attr.Scale)); err == nil {
			if v < maxInteger {
				maxInteger = v
			}
		}

		if minInteger > maxInteger {
			minInteger, maxInteger = maxInteger, minInteger
		}

		fakeVar = fmt.Sprintf(
			"sprintf('%%.%df', fake()%s->numberBetween(%d, %d) / %d)",
			attr.Scale,
			unique,
			minInteger,
			maxInteger,
			factor,
		)
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_LATITUDE) {
			fakeVar = fmt.Sprintf("sprintf('%%.%df', fake()%s->latitude(-90, 90))", attr.Scale, unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_LONGITUDE) {
			fakeVar = fmt.Sprintf("sprintf('%%.%df', fake()%s->longitude(-180, 180))", attr.Scale, unique)
		}
	case proto.AttributeType_ATTRIBUTE_TYPE_STRING, proto.AttributeType_ATTRIBUTE_TYPE_TEXT, proto.AttributeType_ATTRIBUTE_TYPE_LONG_TEXT:
		maxLen := 10
		minLen := 5
		if proto.AttributeType_ATTRIBUTE_TYPE_TEXT == attr.Type {
			minLen = 30
			maxLen = 90
		}
		if proto.AttributeType_ATTRIBUTE_TYPE_LONG_TEXT == attr.Type {
			minLen = 100
			maxLen = 120
		}

		max, err := strconv.Atoi(attr.FakeMax)
		if nil == err {
			maxLen = max
		}
		min, err := strconv.Atoi(attr.FakeMin)
		if nil == err {
			minLen = min
		}
		if maxLen < minLen {
			maxLen = minLen
		}
		if attr.StrFieldLength() > 0 && maxLen > int(attr.StrFieldLength()) {
			maxLen = int(attr.StrFieldLength())
		}
		fakeVar = fmt.Sprintf("fake()%s->lexify(str_repeat('?', rand(%d, %d)))", unique, minLen, maxLen)
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_EMAIL) {
			fakeVar = fmt.Sprintf("fake()%s->safeEmail()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_URL) {
			fakeVar = fmt.Sprintf("fake()%s->url()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_IP) {
			fakeVar = fmt.Sprintf("fake()%s->ipv4()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_ACCOUNT) {
			fakeVar = fmt.Sprintf("fake()%s->userName()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_PASSWORD) {
			fakeVar = "bcrypt('password')"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_NICKNAME) {
			fakeVar = fmt.Sprintf("fake()%s->name()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_AVATAR) {
			fakeVar = "'demos/avatars/' . rand(1, 10) . '.png'"
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_PHONE) {
			fakeVar = fmt.Sprintf("fake()%s->phoneNumber()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_COLOR) {
			fakeVar = fmt.Sprintf("fake()%s->hexColor()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_NAME) {
			fakeVar = fmt.Sprintf("fake()%s->name()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_FIRST_NAME) {
			fakeVar = fmt.Sprintf("fake()%s->firstName()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_LAST_NAME) {
			fakeVar = fmt.Sprintf("fake()%s->lastName()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_USERNAME) {
			fakeVar = fmt.Sprintf("fake()%s->userName()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_SLUG) {
			fakeVar = fmt.Sprintf("fake()%s->slug()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_UUID) {
			fakeVar = fmt.Sprintf("fake()%s->uuid()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_DOMAIN) {
			fakeVar = fmt.Sprintf("fake()%s->domainName()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_IPV6) {
			fakeVar = fmt.Sprintf("fake()%s->ipv6()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_MAC_ADDRESS) {
			fakeVar = fmt.Sprintf("fake()%s->macAddress()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_ADDRESS) {
			fakeVar = fmt.Sprintf("fake()%s->address()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_COUNTRY) {
			fakeVar = fmt.Sprintf("fake()%s->country()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_PROVINCE) {
			fakeVar = fmt.Sprintf("fake()%s->state()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_CITY) {
			fakeVar = fmt.Sprintf("fake()%s->city()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_STREET_ADDRESS) {
			fakeVar = fmt.Sprintf("fake()%s->streetAddress()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_POSTCODE) {
			fakeVar = fmt.Sprintf("fake()%s->postcode()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_LATITUDE) {
			fakeVar = fmt.Sprintf("fake()%s->latitude(-90, 90)", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_LONGITUDE) {
			fakeVar = fmt.Sprintf("fake()%s->longitude(-180, 180)", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_COMPANY) {
			fakeVar = fmt.Sprintf("fake()%s->company()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_JOB_TITLE) {
			fakeVar = fmt.Sprintf("fake()%s->jobTitle()", unique)
		}
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_PHONE) {
			fakeVar = fmt.Sprintf("(string)fake()%s->numberBetween(13000000000, 19999999999)", unique)
		}
	case proto.AttributeType_ATTRIBUTE_TYPE_BOOL:
		fakeVar = "fake()->boolean()"
	case proto.AttributeType_ATTRIBUTE_TYPE_DATETIME:
		fakeVar = fmt.Sprintf("fake()%s->dateTimeBetween('-2 years', 'now')->format('Y-m-d H:i:s')", unique)
	case proto.AttributeType_ATTRIBUTE_TYPE_DATE:
		fakeVar = fmt.Sprintf("fake()%s->date('Y-m-d')", unique)
	case proto.AttributeType_ATTRIBUTE_TYPE_TIME:
		fakeVar = fmt.Sprintf("fake()%s->time()", unique)
	case proto.AttributeType_ATTRIBUTE_TYPE_YEAR:
		fakeVar = fmt.Sprintf("fake()%s->year()", unique)
	case proto.AttributeType_ATTRIBUTE_TYPE_ENUM, proto.AttributeType_ATTRIBUTE_TYPE_SETS:
		if len(attr.Dict().Options) == 0 {
			fakeVar = "null"
		} else {
			fakeVar = fmt.Sprintf("fake()%s->randomElement(%s)", unique, attr.Dict().OptionPhpArray())
		}
	case proto.AttributeType_ATTRIBUTE_TYPE_FILE, proto.AttributeType_ATTRIBUTE_TYPE_FILES:
		fakeVar = "'demos/archives/' . rand(1, 10) . '.zip'"
		if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_AVATAR) {
			fakeVar = "'demos/avatars/' . rand(1, 10) . '.png'"
		} else if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_IMAGE) {
			fakeVar = "'demos/images/' . rand(1, 10) . '.jpg'"
		} else if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_DOCUMENT) {
			fakeVar = "'demos/documents/' . rand(1, 10) . '.pdf'"
		} else if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_ARCHIVE) {
			fakeVar = "'demos/archives/' . rand(1, 10) . '.zip'"
		} else if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_VIDEO) {
			fakeVar = "'demos/videos/' . rand(1, 10) . '.mp4'"
		} else if attr.HasAbility(proto.AttributeAbility_ATTRIBUTE_ABILITY_AUDIO) {
			fakeVar = "'demos/audios/' . rand(1, 10) . '.mp3'"
		}
	}

	if fakeVar == "" {
		panic(fmt.Errorf("no way php fake value use %d attr type", attr.Type))
	}

	if attr.IsFiles() {
		return fmt.Sprintf("array_map(fn () => %s, range(2, 5)),", fakeVar)
	}
	if attr.IsSets() {
		if len(attr.Dict().Options) == 0 {
			return "[]"
		}
		return fmt.Sprintf("array_map(fn () => %s, range(2, 5)),", fakeVar)
	}

	return fmt.Sprintf("%s,", fakeVar)

}

func (attr *Attribute) PhpMigrationToChange() string {
	m := strings.TrimSpace(attr.PhpMigration())
	m = strings.TrimSuffix(m, ";")
	return m + "->change();"
}

func (attr *Attribute) StrFieldLength() int32 {
	//ENUM
	if proto.AttributeType_ATTRIBUTE_TYPE_ENUM == attr.Type {
		return EnumDefaultLen
	}

	//FILE
	if proto.AttributeType_ATTRIBUTE_TYPE_FILE == attr.Type {
		if attr.FieldLength > FileDefaultLen || attr.FieldLength < 1 {
			return FileDefaultLen
		}

		return attr.FieldLength
	}

	if proto.AttributeType_ATTRIBUTE_TYPE_STRING == attr.Type {
		len := attr.FieldLength
		if len > VarCharMaxLen {
			len = VarCharMaxLen
		}
		if len < 1 {
			len = VarCharDefaultLen
		}

		if attr.IsInIndex(false) && len > VarCharIndexDefaultLen {
			len = VarCharIndexDefaultLen
		}

		return len
	}

	return -1
}
