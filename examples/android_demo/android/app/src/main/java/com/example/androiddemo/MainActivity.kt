package com.example.androiddemo

import android.app.Activity
import android.os.Bundle
import androidapp.Androidapp

class MainActivity : Activity() {

    private lateinit var glSurfaceView: GoGuiGLSurfaceView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        Androidapp.init()
        glSurfaceView = GoGuiGLSurfaceView(this)
        setContentView(glSurfaceView)
    }

    override fun onResume() {
        super.onResume()
        glSurfaceView.onResume()
    }

    override fun onPause() {
        glSurfaceView.onPause()
        super.onPause()
    }

    override fun onDestroy() {
        glSurfaceView.queueEvent {
            Androidapp.cleanUp()
        }
        super.onDestroy()
    }
}
