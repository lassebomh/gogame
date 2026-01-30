#version 330

#define MAX_LIGHTS        4
#define LIGHT_DIRECTIONAL 0
#define LIGHT_POINT       1
#define LIGHT_SPOT        2

in vec3 fragPosition;
in vec2 fragTexCoord;
//in vec4 fragColor;
in vec3 fragNormal;

uniform vec4 colDiffuse;
uniform vec4 uvClamp;
uniform sampler2D texture0;

uniform bool fullBright;
uniform sampler2D shadowMap;
uniform vec3 playerPosition;
uniform bool hideOutsideView;

out vec4 finalColor;

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

uniform Light lights[MAX_LIGHTS];
uniform vec4 ambient;
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
  float seenBefore = texture(shadowMap, (((fragPosition.xyz - playerPosition).xz) / (20) + 0.5) * vec2(-1, 1)).r;
  
  
  float viewSample = 0;
  for (float x = -2; x <= 2; x++) {
    for (float y = -2; y <= 2; y++) {
      vec2 uv2 = ((fragPosition.xyz - playerPosition).xz + vec2(x, y)/25) / (20) + 0.5;
      uv2.x *= -1;
      viewSample += texture(shadowMap, uv2).g;
    }
  }
  
  
  float inView = clamp(viewSample / 6 +clamp(5-distance(fragPosition.xyz * vec3(1, 0, 1), playerPosition * vec3(1, 0, 1))*5, 0, 1), 0, 1);
  
  
  float viewDither = 0;
  
  if (inView != 0 && (inView == 1 || (int(gl_FragCoord.x + gl_FragCoord.y)&1) == 0)) {
    viewDither = 1;
  }
  
  float objectInView = 0;

  for (float x = -2; x <= 2; x++) {
    for (float y = -2; y <= 2; y++) {
      vec2 uv2 = ((fragPosition.xyz - playerPosition).xz + vec2(x, y)/5) / (20) + 0.5;
      uv2.x *= -1;
      objectInView += texture(shadowMap, uv2).g / 6;
    }
  }
  
  objectInView += clamp(5-distance(fragPosition.xyz * vec3(1, 0, 1), playerPosition * vec3(1, 0, 1))*5, 0, 1);
  objectInView = clamp(objectInView, 0, 1);
  
  float objectViewDither = 0;
  // float objectViewDither = objectInView;
  
  if (objectInView != 0 && (objectInView == 1 || (int(gl_FragCoord.x + gl_FragCoord.y)&1) == 0)) {
    objectViewDither = 1;
  }
  
  
  float dither = (fract(sin(dot(gl_FragCoord.xy/8, vec2(12.9898, 78.233))) * 43758.5453))*2-1;
  
  vec2 uv = uvClamp.xy + fragTexCoord * (uvClamp.zw - uvClamp.xy);
  vec4 texelColor = texture(texture0, uv);
  
  if (fullBright) {
    finalColor = texelColor * colDiffuse;
    return;
  }
  
  vec3 normal = normalize(fragNormal);
  vec3 viewD = normalize(viewPos - fragPosition);
  
  vec3 lightDot = vec3(0.0);
  vec3 specular = vec3(0.0);

  for (int i = 0; i < MAX_LIGHTS; i++) {
    if (lights[i].enabled == 0) continue;

    if (lights[i].type == LIGHT_DIRECTIONAL) {
    
      vec3 light = -normalize(lights[i].target - lights[i].position);
      float NdotL = max(dot(normal, light), 0.0);
      
      lightDot += lights[i].color.rgb * NdotL * lights[i].strength;

      if (NdotL > 1.0) {
        specular += pow(max(0.0, dot(viewD, reflect(-(light), normal))), 16.0) * lights[i].strength;
      }
      
    } else if (lights[i].type == LIGHT_POINT) {
      
      vec3 light = normalize(lights[i].position - fragPosition);
      float NdotL = max(dot(normal, light), 0.0);
      
      lightDot += lights[i].color.rgb * NdotL * lights[i].strength;

      if (NdotL > 1.0) {
        specular += pow(max(0.0, dot(viewD, reflect(-(light), normal))), 16.0) * lights[i].strength;
      }
      
    } else if (lights[i].type == LIGHT_SPOT) {

      vec3 light = normalize(lights[i].position - fragPosition);
      vec3 rayDir = normalize(lights[i].position - lights[i].target); 
      float theta = dot(light, rayDir); 
      float epsilon = lights[i].cutOff - lights[i].outerCutOff;
      float intensity = clamp((theta - lights[i].outerCutOff) / epsilon, 0.0, 1.0);
      
      

      float NdotL = max(dot(normal, light), 0.0);
      lightDot += (lights[i].color.rgb * NdotL * intensity * lights[i].strength) * inView;

      if (NdotL > 1.0) {
        specular += (pow(max(0.0, dot(viewD, reflect(-light, normal))), 16.0) * intensity) * inView;
      }
    }
  }
  
  
  finalColor = (texelColor*((colDiffuse + vec4(specular, 1.0))*vec4(lightDot, 1.0)));
  // finalColor += texelColor*colDiffuse*clamp((inView), clamp(seenBefore, 0.01, 0.2), .35);
  
  // finalColor *= seenBefore;
  // finalColor.w = 1;
  
  // if (seenBefore > 0 || inView > 0) {
  // } else {
  //   finalColor = vec4(0, 0, 0, 1);
  //   finalColor = (texelColor*((colDiffuse+vec4(specular, 1.0))*vec4(lightDot, 1.0)));
  //   // finalColor += texelColor*colDiffuse*0.01;
  // }

  if (hideOutsideView) {
    finalColor += texelColor*colDiffuse*clamp(objectViewDither * objectInView * 2, 0.08, 0.2);
    finalColor.w = objectViewDither;
  } else {
    finalColor += texelColor*colDiffuse*clamp(inView * viewDither * 2, 0.08, 0.2);    
  }
  
  
  vec3 lab = rgb2lab(finalColor.xyz);
  
  lab.x = floor((lab.x) * 25.0 + dither / 2.0) / 25.0;
  lab.y = floor((lab.y) * 80.0 + dither / 2.0) / 80.0;
  lab.z = floor((lab.z) * 80.0 + dither / 2.0) / 80.0;
    
  finalColor = vec4(
    lab2rgb(lab.xyz),
    finalColor.w
  );
  // finalColor.rgb *= inView;
}

