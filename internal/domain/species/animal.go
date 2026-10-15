package species

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Animal names an animal, or a group of animals, that the application
// can search for and filter by.
type Animal string

// The animals this application supports. It's not everything that the
// GBIF API supports, just a small selection.
const (
	AnimalLion           Animal = "lion"
	AnimalLeopard        Animal = "leopard"
	AnimalCheetah        Animal = "cheetah"
	AnimalElephant       Animal = "elephant"
	AnimalRhinoceros     Animal = "rhinoceros"
	AnimalGiraffe        Animal = "giraffe"
	AnimalZebra          Animal = "zebra"
	AnimalHippopotamus   Animal = "hippopotamus"
	AnimalAfricanBuffalo Animal = "african_buffalo"
	AnimalSpottedHyena   Animal = "spotted_hyena"
	AnimalAfricanWildDog Animal = "african_wild_dog"
	AnimalWildebeest     Animal = "wildebeest"
	AnimalImpala         Animal = "impala"
	AnimalCommonWarthog  Animal = "common_warthog"

	AnimalBlueWhale               Animal = "blue_whale"
	AnimalHumpbackWhale           Animal = "humpback_whale"
	AnimalSpermWhale              Animal = "sperm_whale"
	AnimalFinWhale                Animal = "fin_whale"
	AnimalCommonMinkeWhale        Animal = "common_minke_whale"
	AnimalGreyWhale               Animal = "grey_whale"
	AnimalNorthAtlanticRightWhale Animal = "north_atlantic_right_whale"
	AnimalBowheadWhale            Animal = "bowhead_whale"
	AnimalBeluga                  Animal = "beluga"
	AnimalNarwhal                 Animal = "narwhal"
	AnimalOrca                    Animal = "orca"
	AnimalCommonBottlenoseDolphin Animal = "common_bottlenose_dolphin"
	AnimalCommonDolphin           Animal = "common_dolphin"
	AnimalSpinnerDolphin          Animal = "spinner_dolphin"
	AnimalDuskyDolphin            Animal = "dusky_dolphin"
	AnimalRissosDolphin           Animal = "rissos_dolphin"
	AnimalAmazonRiverDolphin      Animal = "amazon_river_dolphin"

	AnimalAlligator          Animal = "alligator"
	AnimalAmericanAlligator  Animal = "american_alligator"
	AnimalChineseAlligator   Animal = "chinese_alligator"
	AnimalNileCrocodile      Animal = "nile_crocodile"
	AnimalSaltwaterCrocodile Animal = "saltwater_crocodile"
	AnimalAmericanCrocodile  Animal = "american_crocodile"
	AnimalGharial            Animal = "gharial"

	AnimalTiger           Animal = "tiger"
	AnimalJaguar          Animal = "jaguar"
	AnimalSnowLeopard     Animal = "snow_leopard"
	AnimalPuma            Animal = "puma"
	AnimalBrownBear       Animal = "brown_bear"
	AnimalPolarBear       Animal = "polar_bear"
	AnimalGiantPanda      Animal = "giant_panda"
	AnimalRedPanda        Animal = "red_panda"
	AnimalGorilla         Animal = "gorilla"
	AnimalChimpanzee      Animal = "chimpanzee"
	AnimalOrangutan       Animal = "orangutan"
	AnimalRingTailedLemur Animal = "ring_tailed_lemur"

	AnimalRedKangaroo  Animal = "red_kangaroo"
	AnimalKoala        Animal = "koala"
	AnimalPlatypus     Animal = "platypus"
	AnimalCommonWombat Animal = "common_wombat"

	AnimalAmericanBison Animal = "american_bison"
	AnimalMoose         Animal = "moose"
	AnimalSloth         Animal = "sloth"
	AnimalCapybara      Animal = "capybara"

	AnimalEmperorPenguin Animal = "emperor_penguin"
	AnimalKingPenguin    Animal = "king_penguin"
	AnimalGentooPenguin  Animal = "gentoo_penguin"
	AnimalAfricanPenguin Animal = "african_penguin"
	AnimalHarbourSeal    Animal = "harbour_seal"
	AnimalLeopardSeal    Animal = "leopard_seal"
	AnimalWalrus         Animal = "walrus"
	AnimalSeaOtter       Animal = "sea_otter"

	AnimalGreatWhiteShark   Animal = "great_white_shark"
	AnimalWhaleShark        Animal = "whale_shark"
	AnimalGreenTurtle       Animal = "green_turtle"
	AnimalLeatherbackTurtle Animal = "leatherback_turtle"
	AnimalKomodoDragon      Animal = "komodo_dragon"

	AnimalAtlanticPuffin  Animal = "atlantic_puffin"
	AnimalGreaterFlamingo Animal = "greater_flamingo"
	AnimalBaldEagle       Animal = "bald_eagle"
	AnimalSnowyOwl        Animal = "snowy_owl"
	AnimalScarletMacaw    Animal = "scarlet_macaw"
	AnimalShoebill        Animal = "shoebill"
)

// String implements fmt.Stringer on Animal.
func (a Animal) String() string {
	return string(a)
}

// TaxonKeys returns the GBIF Backbone taxon IDs this animal stands for.
// A broad name such as "elephant" covers several species, so this is a
// set.
func (a Animal) TaxonKeys() []int {
	return slices.Clone(animalTaxonKeys[a])
}

// Known reports whether this is an animal the application supports. Prefer
// ParseAnimal where the value came from outside; Known is for a value that
// arrived some other way.
func (a Animal) Known() bool {
	return len(animalTaxonKeys[a]) > 0
}

// ParseAnimal converts a name such as "humpback_whale", "Humpback Whale" or
// "humpback-whale" into a supported Animal, and is the only way an Animal
// should be obtained from user input. An unsupported name is an error rather
// than an Animal nobody can look up.
func ParseAnimal(value string) (Animal, error) {
	name := strings.NewReplacer(" ", "_", "-", "_").Replace(strings.ToLower(strings.TrimSpace(value)))

	animal := Animal(name)
	if !animal.Known() {
		return "", fmt.Errorf("unknown animal %q, run 'workshop animals' to see the options", value)
	}

	return animal, nil
}

// Animals returns every supported animal, in name order.
func Animals() []Animal {
	return slices.Sorted(maps.Keys(animalTaxonKeys))
}

// AnimalOf returns the animal a GBIF taxon key belongs to. Where more
// than one covers it, such as "alligator" and "american_alligator", the
// narrower one wins.
func AnimalOf(gbifKey int) (Animal, bool) {
	a, ok := animalsByKey[gbifKey]
	return a, ok
}

// animalsByKey indexes animalTaxonKeys the other way round, keeping the
// narrower animal where two share a key.
var animalsByKey = func() map[int]Animal {
	index := make(map[int]Animal)
	for _, animal := range Animals() {
		for _, key := range animalTaxonKeys[animal] {
			if prev, ok := index[key]; !ok || len(animalTaxonKeys[animal]) < len(animalTaxonKeys[prev]) {
				index[key] = animal
			}
		}
	}
	return index
}()

// Name is the animal as a reader would write it, such as "Spotted hyena".
func (a Animal) Name() string {
	name := strings.ReplaceAll(string(a), "_", " ")
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

var animalTaxonKeys = map[Animal][]int{
	AnimalLion:           {5219404},                                     // Panthera leo
	AnimalLeopard:        {5219436},                                     // Panthera pardus
	AnimalCheetah:        {2435270},                                     // Acinonyx jubatus
	AnimalElephant:       {2435350, 2435349, 5219461},                   // Loxodonta africana, L. cyclotis, Elephas maximus
	AnimalRhinoceros:     {5220111, 2440880, 5220113, 5220112, 2440878}, // Diceros bicornis, Ceratotherium simum, Rhinoceros unicorns, R. sondaicus, Dicerorhinus sumatrensis
	AnimalGiraffe:        {7716694},                                     // Giraffa
	AnimalZebra:          {2440892, 2440888, 2440894},                   // Equus quagga, E. zebra, E. grevyi
	AnimalHippopotamus:   {2441247},                                     // Hippopotamus amphibius
	AnimalAfricanBuffalo: {2441034},                                     // Syncerus caffer
	AnimalSpottedHyena:   {5218781},                                     // Crocuta crocuta
	AnimalAfricanWildDog: {5219317},                                     // Lycaon pictus
	AnimalWildebeest:     {2441104},                                     // Connochaetes
	AnimalImpala:         {2441144},                                     // Aepyceros melampus
	AnimalCommonWarthog:  {2441212},                                     // Phacochoerus africanus

	AnimalBlueWhale:               {2440735}, // Balaenoptera musculus
	AnimalHumpbackWhale:           {5220086}, // Megaptera novaeangliae
	AnimalSpermWhale:              {8123917}, // Physeter macrocephalus
	AnimalFinWhale:                {2440718}, // Balaenoptera physalus
	AnimalCommonMinkeWhale:        {2440728}, // Balaenoptera acutorostrata
	AnimalGreyWhale:               {2440704}, // Eschrichtius robustus
	AnimalNorthAtlanticRightWhale: {2440339}, // Eubalaena glacialis
	AnimalBowheadWhale:            {2440330}, // Balaena mysticetus
	AnimalBeluga:                  {5220003}, // Delphinapterus leucas
	AnimalNarwhal:                 {5220008}, // Monodon monoceros
	AnimalOrca:                    {2440483}, // Orcinus orca
	AnimalCommonBottlenoseDolphin: {2440447}, // Tursiops truncatus
	AnimalCommonDolphin:           {8324617}, // Delphinus delphis
	AnimalSpinnerDolphin:          {5220060}, // Stenella longirostris
	AnimalDuskyDolphin:            {5220077}, // Lagenorhynchus obscurus
	AnimalRissosDolphin:           {5220027}, // Grampus griseus
	AnimalAmazonRiverDolphin:      {2440314}, // Inia geoffrensis

	AnimalAlligator:          {2441370, 2441368}, // Alligator mississippiensis, A. sinensis
	AnimalAmericanAlligator:  {2441370},          // Alligator mississippiensis
	AnimalChineseAlligator:   {2441368},          // Alligator sinensis
	AnimalNileCrocodile:      {2441341},          // Crocodylus niloticus
	AnimalSaltwaterCrocodile: {2441333},          // Crocodylus porosus
	AnimalAmericanCrocodile:  {2441329},          // Crocodylus acutus
	AnimalGharial:            {2441360},          // Gavialis gangeticus

	AnimalTiger:           {5219416},                   // Panthera tigris
	AnimalJaguar:          {5219426},                   // Panthera onca
	AnimalSnowLeopard:     {2435238},                   // Uncia uncia, which GBIF also knows as Panthera uncia
	AnimalPuma:            {2435099},                   // Puma concolor
	AnimalBrownBear:       {2433433},                   // Ursus arctos
	AnimalPolarBear:       {2433451},                   // Ursus maritimus
	AnimalGiantPanda:      {2433399},                   // Ailuropoda melanoleuca
	AnimalRedPanda:        {5219446},                   // Ailurus fulgens
	AnimalGorilla:         {2436441, 7262070},          // Gorilla gorilla, G. beringei
	AnimalChimpanzee:      {5219534},                   // Pan troglodytes
	AnimalOrangutan:       {5219532, 5707420, 9311132}, // Pongo pygmaeus, P. abelii, P. tapanuliensis
	AnimalRingTailedLemur: {2436412},                   // Lemur catta

	AnimalRedKangaroo:  {5219963}, // Macropus rufus
	AnimalKoala:        {2440012}, // Phascolarctos cinereus
	AnimalPlatypus:     {2433376}, // Ornithorhynchus anatinus
	AnimalCommonWombat: {2440301}, // Vombatus ursinus

	AnimalAmericanBison: {2441176},          // Bison bison
	AnimalMoose:         {2440940},          // Alces alces
	AnimalSloth:         {2436350, 5219520}, // Bradypus, Choloepus hoffmanni
	AnimalCapybara:      {5786666},          // Hydrochoerus hydrochaeris

	AnimalEmperorPenguin: {2481661}, // Aptenodytes forsteri
	AnimalKingPenguin:    {2481660}, // Aptenodytes patagonicus
	AnimalGentooPenguin:  {2481666}, // Pygoscelis papua
	AnimalAfricanPenguin: {5229384}, // Spheniscus demersus
	AnimalHarbourSeal:    {2434793}, // Phoca vitulina
	AnimalLeopardSeal:    {2434790}, // Hydrurga leptonyx
	AnimalWalrus:         {5218819}, // Odobenus rosmarus
	AnimalSeaOtter:       {2433670}, // Enhydra lutris

	AnimalGreatWhiteShark:   {2420694}, // Carcharodon carcharias
	AnimalWhaleShark:        {2417522}, // Rhincodon typus
	AnimalGreenTurtle:       {2442225}, // Chelonia mydas
	AnimalLeatherbackTurtle: {9789983}, // Dermochelys coriacea
	AnimalKomodoDragon:      {2470854}, // Varanus komodoensis

	AnimalAtlanticPuffin:  {2481353}, // Fratercula arctica
	AnimalGreaterFlamingo: {4352332}, // Phoenicopterus roseus
	AnimalBaldEagle:       {2480446}, // Haliaeetus leucocephalus
	AnimalSnowyOwl:        {5959143}, // Bubo scandiacus
	AnimalScarletMacaw:    {5959227}, // Ara macao
	AnimalShoebill:        {2480347}, // Balaeniceps rex
}
