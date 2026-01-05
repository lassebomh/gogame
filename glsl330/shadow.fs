#version 330

out vec4 fragColor;

#define nearPlane 0.01
#define farPlane 1000.0

float linearizeDepth(float d)
{
    float z = d * 2.0 - 1.0;
    return (2.0 * nearPlane * farPlane) /
           (farPlane + nearPlane - z * (farPlane - nearPlane));
}

void main()
{
    float depth = linearizeDepth(gl_FragCoord.z) / farPlane;
    fragColor = vec4(vec3(depth), 1.0);
}
