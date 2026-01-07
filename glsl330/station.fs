#version 330

// Input vertex attributes (from vertex shader)
in vec3 fragPosition;
in vec2 fragTexCoord;
//in vec4 fragColor;
in vec3 fragNormal;



out vec4 finalColor;

uniform vec3 viewPos;

vec3 rgb2lab(vec3 c)
{
    // sRGB to linear
    vec3 rgb = mix(c / 12.92, pow((c + 0.055) / 1.055, vec3(2.4)), step(0.04045, c));

    // Linear RGB to XYZ (D65)
    const mat3 M = mat3(
        0.4124564, 0.3575761, 0.1804375,
        0.2126729, 0.7151522, 0.0721750,
        0.0193339, 0.1191920, 0.9503041
    );
    vec3 xyz = M * rgb;
    xyz /= vec3(0.95047, 1.0, 1.08883); // normalize by reference white

    // XYZ to Lab
    vec3 f = mix(pow(xyz, vec3(1.0 / 3.0)), (7.787 * xyz) + vec3(16.0 / 116.0), step(xyz, vec3(0.008856)));
    float L = 116.0 * f.y - 16.0;
    float a = 500.0 * (f.x - f.y);
    float b = 200.0 * (f.y - f.z);

    // normalize to 0–1
    return vec3(L / 100.0, (a + 128.0) / 255.0, (b + 128.0) / 255.0);
}

vec3 lab2rgb(vec3 c)
{
    // denormalize from 0–1
    float L = c.x * 100.0;
    float a = c.y * 255.0 - 128.0;
    float b = c.z * 255.0 - 128.0;

    float y = (L + 16.0) / 116.0;
    float x = a / 500.0 + y;
    float z = y - b / 200.0;

    vec3 xyz = vec3(x, y, z);
    vec3 xyz3 = pow(xyz, vec3(3.0));
    xyz = mix(xyz3, (xyz - vec3(16.0 / 116.0)) / 7.787, step(xyz3, vec3(0.008856)));

    // Denormalize by reference white
    xyz *= vec3(0.95047, 1.0, 1.08883);

    // XYZ to linear RGB
    const mat3 M = mat3(
         3.2404542, -1.5371385, -0.4985314,
        -0.9692660,  1.8760108,  0.0415560,
         0.0556434, -0.2040259,  1.0572252
    );
    vec3 rgb = M * xyz;

    // linear to sRGB
    rgb = mix(rgb * 12.92, 1.055 * pow(rgb, vec3(1.0/2.4)) - 0.055, step(0.0031308, rgb));

    return clamp(rgb, 0.0, 1.0);
}


void main()
{
    finalColor = vec4(0.0,0.0,0.0,1.0);
    
    float roundoff = 16.0;
    
    
    float dither = fract(sin(dot(gl_FragCoord.xy, vec2(12.9898, 78.233))) * 43758.5453) * 2;
    // float dither = fract(sin(dot(vec3(gl_FragCoord.xyz.xy, (fragPosition.x-fragPosition.z) / 40000.0), vec3(12.9898, 78.233, 37.719))) * 43758.5453) * 2.0;
    dither -= 1;
    
    vec3 lab = rgb2lab(finalColor.xyz);
    
    lab.x = floor((lab.x) * 25.0 + dither / 2.0) / 25.0;
    lab.y = floor((lab.y) * 80.0 + dither / 2.0) / 80.0;
    lab.z = floor((lab.z) * 80.0 + dither / 2.0) / 80.0;
    
    
    // lab.x = floor((lab.x + dither) * 80.0) / 80.0;
    // lab.y = floor((lab.y + dither) * 80.0) / 80.0;
    // lab.z = floor((lab.z + dither) * 80.0) / 80.0;
    
    finalColor = vec4(
      lab2rgb(lab.xyz),
      finalColor.w
    );

}

