package com.example.androiddemo

import android.content.Context
import android.opengl.GLSurfaceView
import android.view.MotionEvent
import androidapp.Androidapp

class GoGuiGLSurfaceView(context: Context) : GLSurfaceView(context) {

    private val density = resources.displayMetrics.density

    init {
        setEGLContextClientVersion(3)
        // 8-bit stencil required for ClipContents.
        setEGLConfigChooser(8, 8, 8, 8, 0, 8)
        setRenderer(GoGuiRenderer(density))
        renderMode = RENDERMODE_CONTINUOUSLY
    }

    override fun onTouchEvent(event: MotionEvent): Boolean {
        val x = event.x / density
        val y = event.y / density
        when (event.action) {
            MotionEvent.ACTION_DOWN -> queueEvent {
                Androidapp.touchBegan(x, y)
            }
            MotionEvent.ACTION_MOVE -> queueEvent {
                Androidapp.touchMoved(x, y)
            }
            MotionEvent.ACTION_UP,
            MotionEvent.ACTION_CANCEL -> queueEvent {
                Androidapp.touchEnded(x, y)
            }
        }
        return true
    }
}
