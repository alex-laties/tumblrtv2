#version 410 core

in vec3 ourColor;
in vec2 TexCoord;

out vec4 color;

uniform sampler2D wat;

void main()
{
  color = texture(wat, TexCoord);
}
