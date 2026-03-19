package com.example.androiddemo

import android.opengl.GLSurfaceView
import androidapp.Androidapp
import javax.microedition.khronos.egl.EGLConfig
import javax.microedition.khronos.opengles.GL10

class GoGuiRenderer(private val density: Float) : GLSurfaceView.Renderer {

    override fun onSurfaceCreated(gl: GL10?, config: EGLConfig?) {
        // Nothing — initialization happens in onSurfaceChanged.
    }

    override fun onSurfaceChanged(gl: GL10?, width: Int, height: Int) {
        val logicalW = (width / density).toInt()
        val logicalH = (height / density).toInt()
        Androidapp.start(logicalW.toLong(), logicalH.toLong(), density)
    }

    override fun onDrawFrame(gl: GL10?) {
        Androidapp.render()
    }
}
