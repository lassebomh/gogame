#version 330

// Input vertex attributes (from vertex shader)
in vec3 fragPosition;
in vec2 fragTexCoord;
//in vec4 fragColor;
in vec3 fragNormal;

// Input uniform values
uniform vec4 colDiffuse;
uniform vec4 uvClamp;
uniform sampler2D texture0;

uniform sampler2D shadowMap;
uniform mat4 playerMvp;
uniform bool fullBright;


// Output fragment color
out vec4 finalColor;

// NOTE: Add here your custom variables

#define     MAX_LIGHTS              4
#define     LIGHT_DIRECTIONAL       0
#define     LIGHT_POINT             1
#define     LIGHT_SPOT              2

#define near 0.01
#define far 5.0

float linearizeDepth(float z)
{
  float d = far - near;
  return ((z - d) / d);
}

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

void main()
{
    // Texel color fetching from texture sampler
    
    vec2 uv = uvClamp.xy + fragTexCoord * (uvClamp.zw - uvClamp.xy);
    vec4 texelColor = texture(texture0, uv);
    
    if (fullBright) {
      finalColor = texelColor * colDiffuse;
      return;
    }
    
    vec3 lightDot = vec3(0.0);
    vec3 normal = normalize(fragNormal);
    vec3 viewD = normalize(viewPos - fragPosition);
    vec3 specular = vec3(0.0);


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
    finalColor += texelColor*(ambient)*colDiffuse;


    // // Shadow calculations
    // vec4 fragPosLightSpace = playerMvp*vec4(fragPosition, 1);
    // fragPosLightSpace.xyz /= fragPosLightSpace.w; // Perform the perspective division
    // fragPosLightSpace.xyz = (fragPosLightSpace.xyz + 1.0)/2.0; // Transform from [-1, 1] range to [0, 1] range
    // vec2 sampleCoords = fragPosLightSpace.xy;
    // float curDepth = fragPosLightSpace.z;

    // // Slope-scale depth bias: depth biasing reduces "shadow acne" artifacts, where dark stripes appear all over the scene
    // // The solution is adding a small bias to the depth
    // // In this case, the bias is proportional to the slope of the surface, relative to the light
    // // float bias = max(0.0002*(1.0 - dot(normal, -viewD)), 0.00002) + 0.00001;
    // float bias = 0.0002;
    // // int shadowCounter = 0;
    // float shadowCounter = 0;
    // // const int numSamples = 9;
    // float sampleDepth = texture(shadowMap, sampleCoords).r;
    // if (curDepth - 0.005 > sampleDepth) {
    //   shadowCounter = 1;
    // };
    // // // PCF (percentage-closer filtering) algorithm:
    // // // Instead of testing if just one point is closer to the current point,
    // // // we test the surrounding points as well
    // // // This blurs shadow edges, hiding aliasing artifacts
    // // vec2 texelSize = vec2(1.0/float(2048));
    // // for (int x = -1; x <= 1; x++)
    // // {
    // //     for (int y = -1; y <= 1; y++)
    // //     {
    // //         float sampleDepth = texture(shadowMap, sampleCoords + texelSize*vec2(x, y)).r;
    // //         // if (curDepth - 0.005 > sampleDepth) shadowCounter++;
    // //         if (curDepth - bias > sampleDepth) shadowCounter++;
    // //     }
    // // }

    

    // finalColor = mix(finalColor, vec4(0, 0, 0, 1), float(shadowCounter));
    // // finalColor = mix(finalColor, vec4(0, 0, 0, 1), float(shadowCounter)/float(numSamples));

    // // Add ambient lighting whether in shadow or not
    // finalColor += texelColor*(ambient/10.0)*colDiffuse;

    // // Gamma correction
    // // finalColor = pow(finalColor, vec4(1.0/2.2));
    
    // // vec4 lightSpacePos = playerMvp * vec4(fragPosition, 1.0);
    // // vec3 projCoords = lightSpacePos.xyz / (lightSpacePos.w);
    // // projCoords = projCoords * 0.5 + 0.5;
    
    // // bool inView = projCoords.x >= 0 && projCoords.x <= 1 && projCoords.y >= 0 && projCoords.y <= 1;
    
    // // if (inView) {
    // //   float closestDepth = texture(shadowMap, projCoords.xy).r;
    // //   float currentDepth = lightSpacePos.z;
    // //   if (currentDepth - 0.005 < closestDepth) {
    // //     // finalColor = vec4(1, 0, 0, 1);
    // //   } else {
    // //     finalColor = vec4(0, 0, 0, 1);
    // //   }
    // // } else {
    // //   finalColor = vec4(0, 0, 0, 1);
    // // }
    // // float shadow = currentDepth - 0.005 > closestDepth ? 1.0 : 0.0;
    // // finalColor = vec4(1, 0, 0, 1);
    // // finalColor.r = (currentDepth / closestDepth) / 5;


}

