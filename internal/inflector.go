package internal

import (
	"io/ioutil"

	"github.com/gedex/inflector"
	"github.com/jinzhu/inflection"
	"gopkg.in/yaml.v2"
)

type Inflector interface {
	Singularize(string) string
	Pluralize(string) string
}

type DefaultInflector struct{}
type RuleInflector struct{}

func (i *DefaultInflector) Singularize(s string) string {
	return inflector.Singularize(s)
}
func (i *DefaultInflector) Pluralize(s string) string {
	return inflector.Pluralize(s)
}

func (i *RuleInflector) Singularize(s string) string {
	return inflection.Singular(s)
}
func (i *RuleInflector) Pluralize(s string) string {
	return inflection.Plural(s)
}

func NewInflector(ruleFile string) (Inflector, error) {
	if ruleFile == "" {
		return &DefaultInflector{}, nil
	}
	err := registerRule(ruleFile)
	if err != nil {
		return nil, err
	}
	return &RuleInflector{}, nil
}

type InflectRule struct {
	Singuler string `yaml:"singular"`
	Plural   string `yaml:"plural"`
}

func registerRule(inflectionRuleFile string) error {
	rules, err := readRule(inflectionRuleFile)
	if err != nil {
		return err
	}
	if rules != nil {
		for _, irr := range rules {
			inflection.AddIrregular(irr.Singuler, irr.Plural)
		}
	}
	return nil
}

func readRule(ruleFile string) ([]InflectRule, error) {
	data, err := ioutil.ReadFile(ruleFile)
	if err != nil {
		return nil, err
	}
	var rules []InflectRule
	err = yaml.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}
	return rules, nil

}
