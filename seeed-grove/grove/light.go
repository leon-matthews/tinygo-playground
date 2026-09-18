package grove

const (
	MaxLux        = 350
	rawSaturation = 44_000
)

// LightSensor is the Grove Light Sensor v1.2.
// A LS06-S phototransistor into a divider + LM358 OpAmp buffer.
// Response time 20-30 milliseconds, peak wavelength is 540nm (green)
// Range is approximately 0..350 lux, where the max is at about 70% of
// the 3.3V rail (a raw reading of about 44,000).
type LightSensor struct {
	adc        ADC
	numSamples uint16
}

// Configure prepares the sensor for use
func (l *LightSensor) Configure(adc ADC, numSamples uint16) {
	if numSamples == 0 {
		panic("zero samples")
	}
	l.adc = adc
	l.numSamples = numSamples
}

// Fraction returns brightness in range 0..1 relative to sensor saturation.
func (l *LightSensor) Fraction() (fraction float32, saturated bool) {
	raw := l.adc.OverSample(l.numSamples)
	clamped := min(raw, rawSaturation)
	if clamped != raw {
		saturated = true
	}
	fraction = float32(clamped) / float32(rawSaturation)
	return fraction, saturated
}

// Lux estimates illuminance from the sensor's nominal 0..350 lux range.
//
// This is quite a low range, suitable only for measuring from darkness to
// typical artificial lighting. The phototransistor is uncalibrated and
// nonlinear; treat as approximate.
func (l *LightSensor) Lux() (lux uint16, saturated bool) {
	fraction, saturated := l.Fraction()
	lux = uint16(fraction*MaxLux + 0.5)
	return lux, saturated
}

// Raw returns value direct from ADC
func (l *LightSensor) Raw() uint16 {
	return l.adc.ReadAnalogValue()
}
