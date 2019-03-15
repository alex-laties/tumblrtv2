//
//  tumblrtv2View.m
//  tumblrtv2
//
//  Created by Alexander Laties on 3/14/19.
//  Copyright © 2019 webscaleaf. All rights reserved.
//

#import "tumblrtv2View.h"
#include <OpenGL/gl.h>
#include <OpenGL/glu.h>
#include <OpenGL/glext.h>
#include <GLUT/glut.h>
#import "go.h"

@implementation tumblrtv2View

- (instancetype)initWithFrame:(NSRect)frame isPreview:(BOOL)isPreview
{
    self = [super initWithFrame:frame isPreview:isPreview];
    if (self) {
        self.glView = [self createGLView];
        [self addSubview:self.glView];
        [self setAnimationTimeInterval:1/30.0];
    }
    return self;
}

- (NSOpenGLView *)createGLView
{
    //This is pretty important.. OS X starts always with a context that only supports openGL 2.1
    //This will ditch the classic OpenGL and initialises openGL 4.1
    NSOpenGLPixelFormatAttribute pixelFormatAttributes[] ={
        NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersion3_2Core,
        NSOpenGLPFAColorSize    , 24                           ,
        NSOpenGLPFAAlphaSize    , 8                            ,
        NSOpenGLPFADoubleBuffer ,
        NSOpenGLPFAAccelerated  ,
        NSOpenGLPFANoRecovery   ,
        0
    };
    NSOpenGLPixelFormat* format = [[NSOpenGLPixelFormat alloc]initWithAttributes:pixelFormatAttributes];
    
    NSOpenGLView* glview = [[NSOpenGLView alloc] initWithFrame:NSZeroRect pixelFormat:format];
    
    NSAssert(glview, @"Unable to create OpenGL view!");
    return glview;
}

- (void)dealloc
{
    [self.glView removeFromSuperview];
    self.glView = nil;
}

- (void)startAnimation
{
    [super startAnimation];
}

- (void)stopAnimation
{
    [super stopAnimation];
}

- (void)drawRect:(NSRect)rect
{
    [super drawRect:rect];
}

- (void)animateOneFrame
{
    [self.glView.openGLContext makeCurrentContext];
    GoGLRender();
    [self.glView update];
    [[self.glView openGLContext] flushBuffer];
    [self setNeedsDisplay:YES];
    return;
}

- (void)setFrameSize:(NSSize)newSize
{
    [super setFrameSize:newSize];
    [self.glView setFrameSize:newSize];
}

- (BOOL)hasConfigureSheet
{
    return NO;
}

- (NSWindow*)configureSheet
{
    return nil;
}

@end
