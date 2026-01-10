#version 330

// Input vertex attributes (from vertex shader)
in vec3 fragPosition;
in vec2 fragTexCoord;
//in vec4 fragColor;
in vec3 fragNormal;

// Input uniform values
uniform sampler2D texture0;
uniform vec4 colDiffuse;

// Output fragment color
out vec4 finalColor;

// NOTE: Add here your custom variables

#define     MAX_LIGHTS              4
#define     LIGHT_DIRECTIONAL       0
#define     LIGHT_POINT             1
#define     LIGHT_SPOT              2

struct MaterialProperty {
    vec3 color;
    int useSampler;
    sampler2D sampler;
};

struct Light {
    int enabled;
    int type;
    vec3 position;
    vec3 target;
    vec4 color;
    
    float cutOff;
    float outerCutOff;
    float strength;
};

// Input lighting values
uniform Light lights[MAX_LIGHTS];
uniform vec4 ambient;
uniform vec3 viewPos;

vec3 rgb2hsv(vec3 c)
{
    vec4 K = vec4(0.0, -1.0 / 3.0, 2.0 / 3.0, -1.0);
    vec4 p = mix(vec4(c.bg, K.wz), vec4(c.gb, K.xy), step(c.b, c.g));
    vec4 q = mix(vec4(p.xyw, c.r), vec4(c.r, p.yzx), step(p.x, c.r));

    float d = q.x - min(q.w, q.y);
    float e = 1.0e-10;
    return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)), d / (q.x + e), q.x);
}

vec3 hsv2rgb(vec3 c)
{
    vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
    vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
    return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}
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
    // Texel color fetching from texture sampler
    vec4 texelColor = texture(texture0, fragTexCoord);
    vec3 lightDot = vec3(0.0);
    vec3 normal = normalize(fragNormal);
    vec3 viewD = normalize(viewPos - fragPosition);
    vec3 specular = vec3(0.0);

    // NOTE: Implement here your fragment shader code

    for (int i = 0; i < MAX_LIGHTS; i++)
    {
        if (lights[i].enabled == 1)
        {
            vec3 light = vec3(0.0);

            if (lights[i].type == LIGHT_DIRECTIONAL)
            {
                light = -normalize(lights[i].target - lights[i].position);
                
                float NdotL = max(dot(normal, light), 0.0);
                lightDot += lights[i].color.rgb * NdotL * lights[i].strength;

                if (NdotL > 1.0) specular += pow(max(0.0, dot(viewD, reflect(-(light), normal))), 16.0) * lights[i].strength; 
            }

            else if (lights[i].type == LIGHT_POINT)
            {
                light = normalize(lights[i].position - fragPosition);
                
                float NdotL = max(dot(normal, light), 0.0);
                lightDot += lights[i].color.rgb * NdotL * lights[i].strength;

                if (NdotL > 1.0) specular += pow(max(0.0, dot(viewD, reflect(-(light), normal))), 16.0) * lights[i].strength; 
            }
            else if (lights[i].type == LIGHT_SPOT)
            {
              vec3 lightDir = normalize(lights[i].position - fragPosition);
    
                vec3 rayDir = normalize(lights[i].position - lights[i].target); 

                float theta = dot(lightDir, rayDir); 
                
                float epsilon = lights[i].cutOff - lights[i].outerCutOff;
                float intensity = clamp((theta - lights[i].outerCutOff) / epsilon, 0.0, 1.0);
                
                light = lightDir;

                float NdotL = max(dot(normal, light), 0.0);
                lightDot += lights[i].color.rgb * NdotL * intensity * lights[i].strength;

                if (NdotL > 1.0) specular += pow(max(0.0, dot(viewD, reflect(-light, normal))), 16.0) * intensity;
            }


        }
    }

    finalColor = (texelColor*((colDiffuse + vec4(specular, 1.0))*vec4(lightDot, 1.0)));
    finalColor += texelColor*(ambient/10.0)*colDiffuse;
    
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

