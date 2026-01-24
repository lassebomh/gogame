#version 330

in float vDepth;
out vec4 fragColor;

#define near 0.0
#define far 100.0


void main()
{
    float depth01 = vDepth; // clamp((vDepth - near) / (far - near), 0.0, 1.0);
    fragColor = vec4(depth01, depth01, depth01, 1.0);
}
