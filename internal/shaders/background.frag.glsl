#version 410 core
out vec4 FragColor;

uniform float alpha;
uniform vec2 cursorPos;
uniform vec2 resolution;
uniform vec3 bgColor;

void main() {
	vec2 fragCoord = gl_FragCoord.xy;
	float dist = length(fragCoord - cursorPos);
	float glowFalloff = smoothstep(0.0, 300.0, dist);
	float cursorTransparency = mix(0.3, 1.0, glowFalloff);

	FragColor = vec4(bgColor, alpha * cursorTransparency);
}
